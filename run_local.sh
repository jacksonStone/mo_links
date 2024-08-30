#!/bin/bash
rm -rf chrome_extension_local/
mkdir chrome_extension_local/
cp -r chrome_extension/* chrome_extension_local/
cp ../product_catalog/product_catalog.json ./server/http_server/static/product_catalog.json
# replace instances of molinks.me with localhost:3003
sed -i '' 's/www.molinks.me/localhost:3003/g' chrome_extension_local/background.js
sed -i '' 's/www.molinks.me/localhost:3003/g' chrome_extension_local/popup.js

sed -i '' 's/"host": "www.molinks.me"/"host": "localhost", "port": "3003"/g' chrome_extension_local/rules.json

sed -i '' 's/Mo Links/Mo Links Local/g' chrome_extension_local/manifest.json

sed -i '' 's/https/http/g' chrome_extension_local/background.js
sed -i '' 's/https/http/g' chrome_extension_local/popup.js
sed -i '' 's/https/http/g' chrome_extension_local/rules.json

cd server && go run .
