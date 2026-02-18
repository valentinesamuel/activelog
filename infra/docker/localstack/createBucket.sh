#!/bin/bash

awslocal s3api head-bucket --bucket activelog-uploads 2>/dev/null

if [ $? -ne 0 ]; then
  echo "Creating bucket..."
  awslocal s3api create-bucket --bucket activelog-uploads
else
  echo "Bucket already exists."
fi
