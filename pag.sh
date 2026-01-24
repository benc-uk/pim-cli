#!/bin/bash

echo -e "🔐 Microsoft Entra ID: PIM for Groups CLI"

# Consts
DURATION="PT720M"  # 12 hours

# Usage
usage() {
  echo "Usage: pag <command> [options]"
  echo ""
  echo "Commands:"
  echo "  list                     List your eligible group roles"
  echo "  status                   Show your active group memberships"
  echo "  activate                 Request group role assignment"
  echo ""
  echo "Options:"
  echo "  list:"
  echo "    -a, --all              List all available group roles in tenant"
  echo ""
  echo "  activate:"
  echo "    -n, --name <name>      Group role name to activate (required)"
  echo "    -r, --reason <reason>  Reason for activation (required)"
  echo ""
  echo "  General:"
  echo "    -h, --help             Show this help message"
  echo ""
  echo "Examples:"
  echo "  $0 list                  List your eligible roles"
  echo "  $0 list --all            List all roles in tenant"
  echo "  $0 status                Show active memberships"
  echo "  $0 activate -n 'My Group' -r 'Need access for deployment'"
  exit 1
}

# Parse command
COMMAND=${1:-}
shift 2>/dev/null || true

# Show help if no command or help requested
if [[ -z "$COMMAND" || "$COMMAND" == "-h" || "$COMMAND" == "--help" || "$COMMAND" == "help" ]]; then
  usage
fi

# Validate command
if [[ "$COMMAND" != "list" && "$COMMAND" != "status" && "$COMMAND" != "activate" ]]; then
  echo "❌ Invalid command: $COMMAND"
  usage
fi

# Parse flags based on command
LIST_ALL=false
ROLE_NAME=""
REASON=""

while [[ $# -gt 0 ]]; do
  case $1 in
    -a|--all)
      LIST_ALL=true
      shift
      ;;
    -n|--name)
      ROLE_NAME="$2"
      shift 2
      ;;
    -r|--reason)
      REASON="$2"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "❌ Unknown option: $1"
      usage
      ;;
  esac
done

# Check if Azure CLI & jq are installed
which az > /dev/null || { echo -e "💥 Error! Azure CLI not installed"; exit 1; }
which jq > /dev/null || { echo -e "💥 Error! jq not installed"; exit 1; }

# Get token and other credential details
TOKEN=$(az account get-access-token --resource https://api.azrbac.mspim.azure.com --query accessToken -o tsv)
USER_OID=$(az ad signed-in-user show --query id -o tsv)
TENANT_ID=$(az account show --query tenantId -o tsv)
USER_NAME=$(az account show --query user.name -o tsv)

if [[ -z "$TOKEN" || -z "$USER_OID" || -z "$TENANT_ID" ]]; then
  echo -e "💥 \e[31mFailed to retrieve access token or user information.\nPlease ensure you are logged in to the Azure CLI.\e[0m"
  exit 1
fi

echo "🏢 Tenant ID: $TENANT_ID"
echo "👷 User Name: $USER_NAME"

# List command
if [[ "$COMMAND" == "list" ]]; then
  if [[ "$LIST_ALL" == true ]]; then
    # List all available roles in the tenant
    ALL_ROLES_RESPONSE=$(curl -s -X GET "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess/aadGroups/resources/" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json")

    echo $ALL_ROLES_RESPONSE

    echo -e "\nAll Available Roles:"
    echo "$ALL_ROLES_RESPONSE" | jq -r '.value[] | "📂 \(.displayName)"' 2>/dev/null

    COUNT=$(echo "$ALL_ROLES_RESPONSE" | jq -r '.value | length' 2>/dev/null)
    if [[ "$COUNT" == "0" || -z "$COUNT" ]]; then
      echo "No roles found."
    fi
  else
    # List eligible assignments for current user
    ELIGIBLE_RESPONSE=$(curl -s -X GET "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess/aadGroups/roleAssignments?\$expand=resource,roleDefinition" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json")

echo $ELIGIBLE_RESPONSE
    echo -e "\nEligible Roles:"
    echo "$ELIGIBLE_RESPONSE" | jq -r '.value[] | "📂 \(.resource.displayName) (\(.roleDefinition.displayName))"' 2>/dev/null
    
    COUNT=$(echo "$ELIGIBLE_RESPONSE" | jq -r '.value | length' 2>/dev/null)
    if [[ "$COUNT" == "0" || -z "$COUNT" ]]; then
      echo "No eligible roles found."
    fi
  fi
  exit 0
fi

# Status command - show active memberships
if [[ "$COMMAND" == "status" ]]; then
  ASSIGNMENTS_RESPONSE=$(curl -s -X GET "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess/aadGroups/roleAssignments?\$filter=subjectId%20eq%20%27${USER_OID}%27%20and%20assignmentState%20eq%20%27Active%27&\$expand=linkedEligibleRoleAssignment,subject,roleDefinition,resource" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

  echo -e "\nActive Memberships:"
  echo "$ASSIGNMENTS_RESPONSE" | jq -r '.value[] | "\(.resource.displayName // "Unknown")|\(.endDateTime // "")"' 2>/dev/null | while IFS='|' read -r name expiry; do
    if [[ -n "$expiry" && "$expiry" != "null" ]]; then
      # Convert ISO date to human-readable format with relative time
      expiry_epoch=$(date -d "$expiry" +%s 2>/dev/null)
      now_epoch=$(date +%s)
      remaining_secs=$((expiry_epoch - now_epoch))
      remaining_mins=$((remaining_secs / 60))
      remaining_hours=$((remaining_mins / 60))
      remaining_mins_mod=$((remaining_mins % 60))
      formatted_date=$(date -d "$expiry" "+%a %d %b %H:%M" 2>/dev/null)
      if [[ $remaining_hours -gt 0 ]]; then
        echo "✅ $name"
        echo "   Expires: $formatted_date (${remaining_hours}h ${remaining_mins_mod}m remaining)"
      elif [[ $remaining_mins -gt 0 ]]; then
        echo "✅ $name"
        echo "   Expires: $formatted_date (${remaining_mins}m remaining)"
      else
        echo "✅ $name"
        echo "   Expires: $formatted_date (expired)"
      fi
    else
      echo "✅ $name"
      echo "   Expires: N/A"
    fi
    echo ""
  done
  
  COUNT=$(echo "$ASSIGNMENTS_RESPONSE" | jq -r '.value | length' 2>/dev/null)
  if [[ "$COUNT" == "0" || -z "$COUNT" ]]; then
    echo "No active memberships found."
  fi
  exit 0
fi


# Function to request a role with given name and reason
request_role() {
  local role_name=$1
  local reason=$2

  echo -e "\n🔐 Requesting: $role_name with reason '$reason'"

  # Get the user's eligible assignments and find the matching role
  local eligible_response
  eligible_response=$(curl -s -X GET "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess/aadGroups/roleAssignments?\$filter=subjectId%20eq%20%27${USER_OID}%27%20and%20assignmentState%20eq%20%27Eligible%27&\$expand=resource,roleDefinition" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

  # Find the resource ID and role definition ID from eligible assignments
  local resource_id role_definition_id
  resource_id=$(echo "$eligible_response" | jq -r --arg name "$role_name" '.value[] | select(.resource.displayName == $name) | .resource.id' 2>/dev/null | head -1)
  role_definition_id=$(echo "$eligible_response" | jq -r --arg name "$role_name" '.value[] | select(.resource.displayName == $name) | .roleDefinition.id' 2>/dev/null | head -1)

  if [[ -z "$resource_id" || -z "$role_definition_id" ]]; then
    echo -e "💥 \e[31mRole '$role_name' not found in your eligible assignments.\e[0m"
    echo "Available roles:"
    echo "$eligible_response" | jq -r '.value[] | "  - \(.resource.displayName)"' 2>/dev/null
    return 1
  fi

  echo "🔄 Activating role..."

  local response
  response=$(curl -s -X POST "https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess/aadGroups/roleAssignmentRequests" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "roleDefinitionId": "'$role_definition_id'",
      "resourceId": "'$resource_id'",
      "subjectId": "'$USER_OID'",
      "assignmentState": "Active",
      "type": "UserAdd",
      "reason": "'"$reason"'",
      "ticketNumber": "",
      "ticketSystem": "",
      "schedule": {
        "type": "Once",
        "startDateTime": null,
        "endDateTime": null,
        "duration": "'$DURATION'"
      }
    }')

  # Parse the response and display status
  local status sub_status
  status=$(echo "$response" | jq -r '.status.status // empty' 2>/dev/null)
  sub_status=$(echo "$response" | jq -r '.status.subStatus // empty' 2>/dev/null)

  if [[ -n "$status" ]]; then
    echo -e "⭐ Status: $status${sub_status:+ ($sub_status)}"
  else
    local error_msg
    error_msg=$(echo "$response" | jq -r '.message // .error.message // empty' 2>/dev/null)
    if [[ -n "$error_msg" ]]; then
      echo -e "❌ Error: $error_msg"
    fi
  fi
}

# Activate command - request role assignments
if [[ "$COMMAND" == "activate" ]]; then
  if [[ -z "$ROLE_NAME" ]]; then
    echo -e "💥 \e[31mPlease provide the group role name with -n or --name.\e[0m"
    usage
  fi
  if [[ -z "$REASON" ]]; then
    echo -e "💥 \e[31mPlease provide a reason with -r or --reason.\e[0m"
    usage
  fi

  request_role "$ROLE_NAME" "$REASON"
  exit $?
fi

