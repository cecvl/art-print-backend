# Web App Server

### Introduction

Back-end for an art prints' sales platform

### Tools

- Go standard libs
- `Cloudinary` for storage
- `Firebase` as backend support, `firestore` as db(obviously lol)

### Curl commands to test server

-`sign up`: curl -X POST http://localhost:3001/signup \
 -H "Content-Type: application/json" \
 -d '{
"email": "useruser@gmail.com",
"password": "securepass123",
"userType": "artist"
}'

- curl -X GET http://localhost:3001/artworks
- curl -X POST http://localhost:3001/cart/add -H "Authorization: Bearer ID_TOKEN" -H "Content-Type: application/json" -d '{
  "artworkId": "7HCcdzetfE69uFLtTfGd",
  "title": "",
  "quantity": 1,
  "price": 1500.00
  }'

- curl -X GET http://localhost:3001/cart -H "Authorization: Bearer ID_TOKEN"
- curl -X POST http://localhost:3001/checkout -H "Authorization: Bearer ID_TOKEN"
