#!/bin/bash
# init-aws.sh
awslocal s3 mb s3://my-local-bucket

awslocal s3api put-bucket-cors --bucket my-local-bucket --cors-configuration '{
  "CORSRules": [
    {
      "AllowedOrigins": ["*"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE", "HEAD"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"]
    }
  ]
}'

echo "Bucket created with CORS."
