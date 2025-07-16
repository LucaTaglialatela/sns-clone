# snsclone-202506-golang-luca

This project is an sns-clone mvp with basic functionalities including CRUD operations on posts, the ability to display, follow or unfollow other users, and more.

## Setup
To get started with this project make sure Go `v1.24.3` and Node `v24.1.0` are installed.

The project uses DynamoDB and S3 for storage. If this is your first time running the project, you might need to initialize your DynamoDB table and S3 bucket. For this there are initializer scripts available. Simply run both of them from the root directory using `go run cmd/initializer/db/init_dynamodb.go` and `go run cmd/initializer/s3/init_s3.go`.

Whether you run the project locally or deploy it somewhere else, e.g. EC2, you need to create a `.env` file in your root directory and populate it with the following variables:
```
// .env
BASE_URL="http://localhost:8000"
PORT="8000"
TABLE_NAME="your-table-name"
BUCKET_NAME="your-bucket-name"
AWS_REGION="your-aws-region"
AWS_ENDPOINT="your-local-aws-endpoint" // If you are running inside EC2 set it to ""
GOOGLE_OAUTH2_CLIENT_ID="your-google-oath2-client-id"
GOOGLE_OAUTH2_CLIENT_SECRET="your-google-oauth2-client-secret"
JWT_SECRET="your-jwt-secret"
```

Also create a `.env.local` file in your frontend directory:
```
// frontend/.env.local
VITE_BASE_URL="http://localhost:8000"
```

If you are running the application locally, for example using LocalStack, please also create a `.envrc` file in your root directory:
```
// .envrc
export AWS_REGION=ap-northeast-1
export AWS_ACCESS_KEY_ID=local
export AWS_SECRET_ACCESS_KEY=fakelocal
unset AWS_SESSION_TOKEN
```

To use LocalStack, you can simply use the provided docker-compose file by running `docker-compose up -d` from the root directory.

## Frontend
The frontend is served as static files through the backend. You might need to first run `npm install` to install all the dependencies. Then, to generate the frontend navigate into the frontend folder and run `npm run build`. This wil automatically create a `dist` folder which will be served by the backend.

## Backend
After generating the frontend files, run the backend from the root directory using `go run cmd/api/main.go`. Your application should now be accessible at the specified `HOST:PORT` address.

## Other
This repo also provides a Caddyfile if you want to use caddy as a reverse proxy for https. Make sure to update the base url environment variables to include https. Additionally, systemd service files are provided to launch the application (and caddy) on system startup. It is assumed you have installed caddy and set up your application binary. To do this navigate to the project root directory and create the binary using `go build sns-clone cmd/api/main.go` then move it and the `.env` file to `/srv/sns-clone`.
