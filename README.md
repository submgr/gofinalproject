# Classifieds Platform

A modern classified advertisements platform built with Go and Tailwind CSS.

## Features

- User authentication and authorization
- Advertisement management
- Category system
- Image upload support
- Search and filtering
- Responsive design with Tailwind CSS

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Project Structure

```
.
├── backend/
│   ├── api/         # API handlers
│   ├── models/      # Database models
│   ├── database/    # Database configuration
│   └── middleware/  # Middleware functions
└── frontend/
    ├── static/      # Static assets
    └── templates/   # HTML templates
```

## Setup and Running

1. Clone the repository:
```bash
git clone <repository-url>
cd classifieds
```

2. Install dependencies:
```bash
cd backend
go mod tidy
```

3. Create a `.env` file in the backend directory:
```bash
PORT=8080
JWT_SECRET=your-secret-key
```

4. Run the application:
```bash
go run main.go
```

The application will be available at `http://localhost:8080`

## API Endpoints

- `GET /api/health` - Health check endpoint
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/advertisements` - List advertisements
- `POST /api/advertisements` - Create advertisement
- `GET /api/advertisements/:id` - Get advertisement details
- `PUT /api/advertisements/:id` - Update advertisement
- `DELETE /api/advertisements/:id` - Delete advertisement

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request 