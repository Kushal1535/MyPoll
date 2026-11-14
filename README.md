\# MyPoll – Real-Time Live Polling Application



MyPoll is a real-time polling application that allows users to create polls, share public links, vote, and view live results without refreshing the page.



\## Features

\- User registration and login

\- JWT authentication

\- Create and manage polls

\- Public poll sharing and voting

\- Duplicate-vote prevention

\- Live updates using WebSockets

\- Redis Pub/Sub for real-time event distribution

\- MongoDB data storage

\- Live dashboard results

\- Protected poll creation and ownership



\## Technology Stack



\*\*Frontend:\*\* React, TypeScript, Vite, React Router



\*\*Backend:\*\* Go, Gin, JWT, WebSockets



\*\*Database and Realtime:\*\* MongoDB, Redis



\## Project Structure

\- `backend/` – Go backend application

\- `frontend/` – React frontend application

\- `.gitignore` – Git ignore rules

\- `README.md` – Project documentation



\## Run Locally



\### Prerequisites

Install Go, Node.js, MongoDB, and Redis.



\### Backend

1\. Start MongoDB and Redis.

2\. Open a terminal in the backend folder.

3\. Set the JWT\_SECRET environment variable.

4\. Run `go run .`



Backend: `http://localhost:8081`



\### Frontend



Open another terminal:



```bash

cd frontend

npm install

Copy-Item .env.example .env

npm run dev

```



Frontend: `http://localhost:5173`



\## Health Check Endpoints

\- `GET /api/health`

\- `GET /api/health/db`

\- `GET /api/health/redis`



\## Real-Time Architecture



When a user votes, the backend stores the vote in MongoDB, updates Redis, and publishes an event through Redis Pub/Sub. WebSocket connections deliver updates to connected clients without refreshing the page.



\## Security

\- Do not commit `.env` files or credentials.

\- Use a unique JWT secret in each environment.

\- Configure production database credentials and allowed frontend origins before deployment.



\## Author

MyPoll – Real-Time Live Polling Application

