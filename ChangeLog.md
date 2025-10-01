# ChangeLog

Based on git history and code analysis, here's the ChangeLog for the Symbio freelance platform.

## Recent Changes (Latest Commits)

- **8fcd32b**: Fixed regexp for amount validation to ensure correct amounts processing.
- **fec9dd9**: Fixed maximum password size validation.
- **9cc202e**: Prevented blocked users from creating chat requests.
- **a4714e2**: Prevented duplicate chat requests and updated Swagger documentation.
- **21eeefe**: Fixed wallet creation logic after Grok integration.
- **e57388b**: Fixed create wallet logic.
- **4669910**: Fixed database migration issues.
- **1c13150**: Fixed escrow balance transfers and improved rate limiting.
- **b7962c4**: Added design updates by Grok.
- **1574cb6**: Added minimal amount validation for sends.
- **ae22073**: Added rating system for users.
- **b75d1d9**: Added pagination for user profiles.
- **bf9176d**: Centralized configuration and improved code structure.
- **09abd93**: Added feature to hide blocked user profiles.

## Major Features Implemented

- **User Authentication & Management**: JWT-based auth, CAPTCHA protection, password recovery with mnemonic phrases.
- **Task Management**: Create, update, delete tasks; offer system; escrow integration.
- **Wallet Integration**: Bitcoin (Electrum) support with balance management and transactions.
- **Dispute Resolution**: Admin-managed dispute system with messaging.
- **Chat System**: User-to-user chat with request/accept flow.
- **Admin Panel**: User moderation, transaction monitoring, balance adjustments, permissions system.
- **Review System**: Task completion reviews with rating.
- **Support Tickets**: User support system with admin assignment.
- **Escrow System**: Secure fund holding for tasks.
- **Lua Scripting**: Extensible backend with Lua mods and WASM support.
- **API Documentation**: Comprehensive Swagger docs.
- **Web Frontend**: Symfony-based PHP web interface.
- **Prototypes**: React single-file app and HTML prototypes for UI concepts.

## Database Schema

- Complete PostgreSQL schema with tables for users, profiles, wallets, tasks, offers, transactions, escrow, disputes, reviews, chat, tickets.
- Redis for caching and session management.

## Security Features

- Rate limiting, input validation, SQL injection protection, XSS prevention, user blocking.