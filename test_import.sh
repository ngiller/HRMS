#!/bin/bash
TOKEN=$(curl -s -X POST http://localhost:8900/api/auth/login -H "Content-Type: application/json" -d '{"email":"dewi.sartika@company.com","password":"password123"}' | jq -r '.data.access_token')
echo "Token: $TOKEN"
curl -s -X POST http://localhost:8900/api/employees/import -H "Authorization: Bearer $TOKEN" -F "file=@employee_template.xlsx" | jq .
