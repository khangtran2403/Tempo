# Tempo - Open Source Workflow Automation Platform

Tempo is a powerful, open-source workflow automation platform designed to help you connect your favorite apps and services, automating repetitive tasks without writing complex code. Think of it as a self-hosted alternative to Zapier or IFTTT, giving you full control over your data and infrastructure.

## 🚀 Key Features

*   **Visual Workflow Builder:** Intuitive drag-and-drop interface powered by React Flow for designing complex automation sequences.
*   **Flexible Triggers:**
    *   **Webhook:** Trigger workflows instantly from external applications via HTTP POST requests.
    *   **Cron:** Schedule workflows to run at specific times or recurring intervals.
*   **Diverse Actions (Connectors):**
    *   **HTTP:** Send custom HTTP requests (GET, POST, PUT, DELETE) to any API.
    *   **Email:** Send automated emails via SMTP.
    *   **Excel:** Generate `.xlsx` report files dynamically from workflow data.
    *   **Google Integration:** Upload files to **Google Drive**, append rows to **Google Sheets**, and upload objects to **Google Cloud Storage (GCS)**.
    *   **GitHub:** Automate developer tasks like creating Issues, Pull Requests, or adding comments.
    *   **Notion:** Create pages and manage databases within your Notion workspace.
    *   **Discord:** Send messages and rich embeds to Discord channels.
*   **Integration Management:** Securely connect and manage third-party accounts (Google, GitHub, Notion) using OAuth2.
*   **Dynamic Templating:** Powerful Go-based templating engine to pass data seamlessly between workflow steps (e.g., `{{ .trigger.data.email }}`). Support for custom Node IDs makes referencing data intuitive.
*   **Execution History & Observability:** Track the success, failure, and execution time of every workflow run. View detailed input and output payloads for easy debugging.
*   **Secrets Management:** Securely store and use sensitive information (API keys, tokens) within your workflows without exposing them in plain text.
*   **Robust Architecture:** Built on top of [Temporal.io](https://temporal.io/) to guarantee workflow execution reliability, retries, and state management, even in the face of infrastructure failures.

## 🛠️ Technology Stack

*   **Backend:** Go, Gin Framework, GORM
*   **Orchestration Engine:** Temporal.io
*   **Database:** PostgreSQL
*   **Frontend:** React, TypeScript, Tailwind CSS, React Flow, React Query, Zustand

## 📦 Getting Started

### Prerequisites

*   [Go](https://golang.org/doc/install) (1.20+)
*   [Node.js](https://nodejs.org/) & npm/yarn
*   [Docker](https://docs.docker.com/get-docker/) & Docker Compose (for running Temporal and PostgreSQL)

### 1. Infrastructure Setup (Temporal & Database)

It is highly recommended to run Temporal and PostgreSQL using Docker Compose.

```bash
# Clone the repository
git clone <your-repo-url> tempo
cd tempo

# Start Temporal and PostgreSQL services (Assuming you have a docker-compose.yml setup)
# docker-compose up -d
```

### 2. Backend Setup

1.  Navigate to the project root.
2.  Create a `.env` file based on `.env.example` (or configure the following variables):
3.  Start the Tempo API Server:
```bash
go run ./cmd/api/main.go
```

4.  Open a new terminal window and start the Tempo Worker (Crucial for executing workflows):
```bash
go run ./cmd/worker/main.go
```

### 3. Frontend Setup

1.  Navigate to the `web` directory:
```bash
cd web
```

2.  Install dependencies:
```bash
npm install
```

3.  Start the development server:
```bash
npm start
```

The application will be available at `http://localhost:3000`.

## 📖 Usage Examples

### Webhook to Discord Notification
1. Create a new Workflow.
2. Add a **Webhook** trigger.
3. Add a **Discord** action. Configure the Action to `send_message`, provide your Discord channel webhook URL, and use templating in the message content, e.g., `New alert received: {{ .trigger.data.message }}`.
4. Save and activate the workflow.
5. Send a POST request to the generated Webhook URL with a JSON payload `{"message": "Hello World!"}`.

### Scheduled Google Sheets Export
1. Create a new Workflow.
2. Add a **Cron** trigger (e.g., `0 8 * * *` for 8 AM daily).
3. Add an **HTTP** action to fetch data from your API. Give this node an ID like `fetch_data`.
4. Add a **Google Sheets** action. Connect your Google account, provide the Spreadsheet ID, and configure the Row Data using the output from the HTTP action: `{{ .fetch_data.body.items | json }}`.
5. Save and activate.

## 🤝 Contributing
Contributions are welcome! Please feel free to submit a Pull Request.
