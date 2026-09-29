#!/bin/bash

# A script to automatically verify AWS OIDC and IAM Role configurations for a project.

PROJECT_NAME=$1

if [ -z "$PROJECT_NAME" ]; then
    echo "Usage: ./verify-aws-setup.sh <project-name>"
    exit 1
fi

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text 2>/dev/null)
if [ $? -ne 0 ]; then
    echo "❌ Failed to get AWS Account ID. Are you logged in?"
    exit 1
fi

OIDC_ARN="arn:aws:iam::${ACCOUNT_ID}:oidc-provider/token.actions.githubusercontent.com"
ROLE_NAME="${PROJECT_NAME}-github-actions-role"

echo "🔍 Verifying OIDC Provider..."
OIDC_INFO=$(aws iam get-open-id-connect-provider --open-id-connect-provider-arn "$OIDC_ARN" 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "✅ OIDC Provider exists."
    # Extract thumbprints
    THUMBPRINTS=$(echo "$OIDC_INFO" | grep -o '"[a-f0-9]\{40\}"' | tr -d '"')
    echo "   Thumbprints found:"
    for tp in $THUMBPRINTS; do
        echo "   - $tp"
    done
else
    echo "❌ OIDC Provider does not exist in AWS Account ${ACCOUNT_ID}."
fi

echo ""
echo "🔍 Verifying GitHub Actions IAM Role..."
ROLE_ARN=$(aws iam get-role --role-name "$ROLE_NAME" --query 'Role.Arn' --output text 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "✅ IAM Role exists: $ROLE_ARN"
    
    # Check trust policy
    echo "   Trust Policy Sub Claims (Repositories allowed):"
    aws iam get-role --role-name "$ROLE_NAME" --query 'Role.AssumeRolePolicyDocument.Statement[0].Condition.StringLike."token.actions.githubusercontent.com:sub"' --output json | grep -v 'null' || echo "   (Could not parse sub claims)"
else
    echo "❌ IAM Role '$ROLE_NAME' does not exist."
fi
