#!/bin/bash

read -p "Registry [https://index.docker.io/v2/]: " registry
registry=${registry:-https://index.docker.io/v2/}
read -p "Username: " username
read -sp "Password: " password
echo
read -p "Email: " email

echo "{
  \"auths\": {
    \"${registry}\": {
      \"username\": \"${username}\",
      \"password\": \"${password}\",
      \"email\": \"${email}\",
      \"auth\": \"$(echo -n "$username:$password" | base64)\"
    }
  }
}" > ./config/manager/secrets/.dockerconfigjson
