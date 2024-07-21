# Step 1: Build the Go binary
echo "Building Go binary..."
GOOS=linux GOARCH=amd64 go build  -o mo_links ./server/server.go|| { echo "Go build failed"; exit 1; }
# Step 2: SCP the Go binary to the EC2 instance
echo "Copying Go binary to EC2 instance..."
scp -i $EC2_PEM_PATH mo_links ubuntu@$EC2_PUBLIC_IP:/home/ubuntu/.temp/ || { echo "SCP failed"; exit 1; }

echo "Connecting to EC2 instance and moving the file..."
ssh -i $EC2_PEM_PATH ubuntu@$EC2_PUBLIC_IP << EOF
  mv ./.temp/mo_links . || { echo "Failed to move the file"; exit 1; }
  chmod +x mo_links || { echo "Failed to make the file executable"; exit 1; }
  echo "File moved successfully"
  sudo systemctl restart mo_links || { echo "Failed to restart"; exit 1; }
EOF