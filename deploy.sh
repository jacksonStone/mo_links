#!/bin/bash
cp ../product_catalog/product_catalog.json ./server/http_server/static/product_catalog.json

# Step 2: SCP the Go binary to the EC2 instance
echo "Copying Go binary to EC2 instance..."
scp -i $EC2_PEM_PATH -r server ubuntu@$EC2_PUBLIC_IP:/home/ubuntu/.temp || { echo "SCP failed"; exit 1; }

echo "Connecting to EC2 instance and moving the file..."
ssh -i $EC2_PEM_PATH ubuntu@$EC2_PUBLIC_IP << EOF
    cd .temp/server
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 /usr/local/go/bin/go build  -o ./../../mo_links . || { echo "Go build failed for server"; exit 1; }
    cd ../../
    sudo chmod +x mo_links
    rm -rf .temp
    mkdir .temp
    sudo systemctl restart mo_links
EOF