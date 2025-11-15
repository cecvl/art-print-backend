# Web App Server

### Introduction

Back-end for an art prints' sales platform

### Tools

- Go standard libs
- `Cloudinary` for storage
- `Firebase` as backend support, `firestore` as db

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

# Make commands
Run `make --version` to check installation
##
`make run` - run app server

`make build` - build a binary file

`make build-win`
`make build-linux` - build for windows, linux

`make seed` - seed json data

`make emulators` - run firebase emulators

