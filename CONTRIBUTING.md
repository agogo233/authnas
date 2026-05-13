# Contributing to AuthNas

Contributions to AuthNas are welcome! Lets get started.

## Support

Issues, Suggestions, and Feature Requests should be added as [Issues](https://github.com/authnas/authnas/issues) of the appropriate type. For Help and Support, Q&A, or anything else; open a [Discussion](https://github.com/orgs/authnas/discussions).

## Documentation Updates

Documentation updates should made in the docs/ directory and a PR opened for approval of the changes.

## Development Environment

AuthNas consists of a Frontend (`./web`) and Backend (`./go-server`). The frontend is served through the backend, and when developing the frontend is automatically rebuilt when changes are detected. To see those frontend changes on the web UI the page must be refreshed. Lets get set up!

### Prerequisites

- An IDE supporting Go development, such as VSCode
- Go >= 1.25 installed on your development machine
- Docker Compose, or different way to run your own local Postgres DB

### Starting the Project

All paths and actions are in the project root directory unless otherwise specified.

1. Navigate to `./go-server` and run `go mod download` to install Go dependencies.
2. Configure a `.env` file using the `.example.env` file as a template in the `go-server` directory. All variables in `.example.env` are required in `.env`, though you can add additional variables as well.
3. Run `docker compose up -d authnas-db` or equivalent in your terminal to start the AuthNas database locally.
4. From the `./go-server` directory, run `go run ./cmd/server` to start the backend server.
5. For the frontend, navigate to `./web` and run `npm install` to install frontend dependencies.
6. From the `./web` directory, run `npm run dev` to start the frontend development server.
7. Visit `localhost:5173` or your configured `APP_URL` to view the AuthNas web UI.

## Contribution Standards

All code contributions are expected to be thoroughly tested and defect-free. Changes must also pass linting for frontend code, which are checked automatically for every PR. You can manually run linting locally by running `npm run lint` in the web directory.

For Go backend code, ensure `go test ./...` passes before submitting PR.

Pull Requests (PRs) should follow the template that appears when they are opened. Make sure to list every feature and fix contained in the PR, screenshots should be attached if any changes to the frontend UI were made.
