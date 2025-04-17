# Complaints API

Backend API for submitting and managing complaints.  
Each user can submit only one active complaint until it's reviewed (approved or rejected).  
Authentication is done via JWT. Built using Go, chi, PostgreSQL and Redis.

## Features

- Submit a new complaint (one per user at a time)
- Admin review system: approve or reject complaints
- JWT-based authentication
- Middleware for user verification
- PostgreSQL as the primary database
- Redis for caching or limiting duplicate complaint submissions

## Tech Stack

- **Language:** Go
- **Framework:** chi
- **Auth:** JWT
- **Database:** PostgreSQL
- **Cache / Locking:** Redis
- **Others:** Docker (optional), Makefile

## Installation

```bash
git clone https://github.com/mdqni/complaints_api.git
cd complaints_api
go mod tidy
