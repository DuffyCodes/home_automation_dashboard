
# Quick Stats API

An API to ingest, store, and display quick stats using **Grafana**, **Prometheus**, **MongoDB**, and **Go**.

## Features

- **Ingest Data**: Accepts incoming data through HTTP endpoints for seamless integration.
- **Store Data**: Leverages **MongoDB** for efficient and scalable data storage.
- **Monitor with Prometheus**: Exposes metrics for Prometheus to track system performance and health.
- **Visualize with Grafana**: Provides insightful dashboards and visualizations the data.

## Technology Stack

- **Go**: Backend API implementation.
- **MongoDB**: Database for storing stats data.
- **Prometheus**: Metrics collection and monitoring.
- **Grafana**: Visualization and dashboarding tool.

## Requirements
- Home Assistant installed and properly set up (Raspberry Pi, VM, etc)
- MQTT configured and working (mosquitto or something similar for a broker)
- Information on connecting to your Home Assistant instance

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/DuffyCodes/home_automation_dashboard.git
   cd home_automation_dashboard
   ```

2. Build the Go application:
   Will fill out when i finalize how it will work

3. Start MongoDB, Prometheus, and Grafana using your preferred method (e.g., Docker or local installation).

4. Run the API:
   Will fill out when i finalize how it will work

5. Access the API and dashboards:
   - API: `http://localhost:8080`
   - Grafana: `http://localhost:3000`
