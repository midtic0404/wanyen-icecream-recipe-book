# 🍦 Wanyen Ice Cream Recipe Book

A simple web application for managing ice cream recipes with full CRUD functionality.

## Features

- **SQLite Database**: Persistent storage for recipes with fields: id, name, ingredients, instructions, prep_time, created_at
- **HTMX Integration**: Smooth, real-time interactions without page reloads
- **Full CRUD Operations**: Create, read, update, and delete recipes
- **Responsive Design**: Modern, mobile-friendly interface
- **Modal Forms**: Clean UI for adding and editing recipes

## Tech Stack

- **Backend**: Go with built-in HTTP server
- **Database**: SQLite with go-sqlite3 driver
- **Frontend**: HTML templates with HTMX for dynamic interactions
- **Styling**: Custom CSS with gradient design

## Quick Start

1. **Clone and run**:
   ```bash
   git clone <repo-url>
   cd icecream-recipe-book
   go mod tidy
   go run main.go
   ```

2. **Access the app**: Open http://localhost:8080

## Usage

- **View recipes**: See all recipes on the home page
- **Add recipe**: Click "Add New Recipe" button
- **View details**: Click "View Details" on any recipe card
- **Edit recipe**: Click "Edit Recipe" in the detail modal
- **Delete recipe**: Click "Delete Recipe" with confirmation

The app automatically creates sample recipes on first run.