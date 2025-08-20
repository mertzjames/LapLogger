# LapLogger

A comprehensive swim timing tracker application for monitoring stroke performance across different events.

## Features

- **User Authentication**: Secure user registration and login with JWT tokens
- Track multiple swimming strokes (Freestyle, Backstroke, Breaststroke, Butterfly)
- Support for various swimming events (50m, 100m, 200m, etc.)
- Time logging and performance analytics
- Clean, responsive web interface
- RESTful API backend with protected routes

## Tech Stack

- **Backend**: Go with Gorilla Mux router and SQLite database
- **Frontend**: React.js with modern hooks and responsive design
- **Database**: SQLite for local data storage

## Project Structure

```
LapLogger/
├── backend/           # Go API server
│   ├── main.go
│   ├── models/
│   ├── handlers/
│   └── database/
├── frontend/          # React.js application
│   ├── src/
│   ├── public/
│   └── package.json
└── README.md
```

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Node.js 18 or higher
- Git

### Quick Start
1. Clone the repository:
   ```bash
   git clone https://github.com/mertzjames/LapLogger.git
   cd LapLogger
   ```

2. Run the application:
   ```bash
   ./start.sh
   ```

   This will start both the backend (port 8080) and frontend (port 3000).

### Manual Setup

#### Backend Setup
```bash
cd backend
go mod tidy
go run .
```
The API server will start on http://localhost:8080

#### Frontend Setup
```bash
cd frontend
npm install
npm start
```
The React app will start on http://localhost:3000

## Usage

1. **Add Swimmers**: Start by adding swimmers to the system
2. **Log Times**: Record swim times for different events and strokes
3. **View Dashboard**: See statistics and recent times
4. **Browse Times**: Filter and view all recorded times

## Database

The application uses SQLite for data storage. The database file (`laplogger.db`) is created automatically in the backend directory when you first run the server.

### Pre-loaded Data

The application comes with pre-configured:
- Swimming strokes (Freestyle, Backstroke, Breaststroke, Butterfly, Individual Medley)
- Common events (50m, 100m, 200m, etc. for each stroke)

## API Endpoints

- `GET /api/swimmers` - Get all swimmers
- `POST /api/swimmers` - Create new swimmer
- `GET /api/times/:swimmer_id` - Get times for a swimmer
- `POST /api/times` - Log new time
- `GET /api/events` - Get all events
- `GET /api/strokes` - Get all stroke types

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request
