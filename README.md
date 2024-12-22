# Dev Tool
Dev Tool is a Go CLI application which simplifies testing webhooks and hosting static websites using dynamically created Kubernetes resources.
## Description
This project provides on-demand temporary servers for testing webhooks or hosting static sites. It communicates with a central server that dynamically creates Peripheral Servers which users can use to relay incoming requests or serve static content without needing to manage complex networking or domain configuration.

My ```Azure Free Trial expires at 06/01/2025```, so my application will not work after that.

## How it Works
1. The user runs a CLI (command-line interface) application.
2. The CLI requests a peripheral server(temporary environment) from the central server.
3. The central server programmatically creates:
   - A Pod (container running your peripheral server)
   - A Service (exposing the Pod inside the Kubernetes cluster)
   - An Ingress (providing an external endpoint)
   - A DNS entry (allowing a custom subdomain to direct traffic to the Pod)
4. The user is given a unique subdomain using which they can reach their temporary server over the public internet.
5. After approximately one hour, the central server cleans up the resources, ending the temporary session.

## Usage
This project can be used for:

- Webhook or API endpoint testing.
- Static website demonstration or temporary hosting.

### Webhook Testing
1. Build and run the local application that on which you want to receive incoming webhook data:
2. When you run CLI application with  Option "1" (Webhook Testing) it prompts you for the local application address (e.g., http://localhost:5000/requests).
3. The CLI then asks the central server for a temporary remote endpoint (subdomain). Any requests sent to that subdomain are forwarded to your local server.
4. For external integrations (e.g., GitHub or Stripe webhooks), set the webhook URL to the temporary subdomain you received.

### Static Website Hosting
1. Place your static files in the static directory (or configure a custom path).
2. Run the CLI application with option "2" (Website Demo). The CLI requests the central server to provide a temporary subdomain.
3. Once you receive the subdomain transfer static website files to the subdomain similar to local_app example.
4. Access your temporary website using the provided subdomain. Share the URL with others to quickly preview or test changes.
