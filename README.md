# LapLogger

A comprehensive swim timing tracker application for monitoring stroke performance across different events.

## Features

- Track multiple swimming strokes (Freestyle, Backstroke, Breaststroke, Butterfly)
- Support for various swimming events (50m, 100m, 200m, etc.)
- Time logging and performance analytics
- Clean, responsive web interface
- RESTful API backend

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

### Backend Setup
```bash
cd backend
go mod init laplogger
go run main.go
```

### Frontend Setup
```bash
cd frontend
npm install
npm start
```

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
