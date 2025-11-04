#!/bin/bash
# https://docs.localstack.cloud/references/init-hooks/#lifecycle-stages-and-hooks
aws configure set region ap-northeast-1 
aws configure set output json

aws --endpoint-url http://localhost:4566 ssm put-parameter \
    --name '/app_name1/local/service_name1/DB_HOST' \
    --type 'SecureString' \
    --value $DB_HOST

aws --endpoint-url http://localhost:4566 ssm put-parameter \
    --name '/app_name1/local/service_name1/DB_PORT' \
    --type 'SecureString' \
    --value $DB_PORT

aws --endpoint-url http://localhost:4566 ssm put-parameter \
    --name '/app_name1/local/service_name1/DB_USER' \
    --type 'SecureString' \
    --value $DB_USER

aws --endpoint-url http://localhost:4566 ssm put-parameter \
    --name '/app_name1/local/service_name1/DB_PASSWORD' \
    --type 'SecureString' \
    --value $DB_PASSWORD

aws --endpoint-url http://localhost:4566 ssm put-parameter \
    --name '/app_name1/local/service_name1/DB_NAME' \
    --type 'SecureString' \
    --value $DB_NAME

# TODO: --secret-stringもdocker-compose.ymlのenvironmentで指定した値を使う
aws --endpoint-url http://localhost:4566 secretsmanager create-secret \
    --name 'test-secret1' \
    --description 'test secret 1' \
    --secret-string '{"SECRET_VALUE_1":"value1","SECRET_VALUE_2":"value2"}'

aws --endpoint-url http://localhost:4566 secretsmanager create-secret \
    --name 'test-secret2' \
    --description 'test secret 2' \
    --secret-string '{"SECRET_VALUE_3":"value3","SECRET_VALUE_4":"value4"}'

echo "Initialized for the Ensoria project!"