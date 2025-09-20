Here is a complete, end-to-end guide for the entire project, designed for a beginner. I've included detailed troubleshooting steps for the most common errors you might encounter along the way.

## 🚀 End-to-End DevOps Project: CI/CD for a Go Application

This guide will walk you through building a complete CI/CD pipeline from scratch. We will write a simple application, containerize it with Docker, set up a CI pipeline with GitHub Actions, and deploy it to a Kubernetes cluster using an ArgoCD GitOps workflow.

### 🏛️ Architecture

```
Developer --> Git Push --> GitHub --> GitHub Actions (Builds Image) --> Docker Hub --> ArgoCD (Detects Change) --> Deploys to Kubernetes Cluster --> User
```

-----

### Stage 1: The Foundation (Azure VM & Tooling)

#### 1\. Create the Azure Virtual Machine

First, we need our server in the cloud.

  * Log in to the Azure Portal.
  * Click **Create a resource** and search for **Virtual machine**.
  * **Basics Tab**:
      * **Image**: **Ubuntu Server 22.04 LTS**.
      * **Size**: Select a size with at least **2 vCPUs and 8 GiB of RAM**. A good choice is `Standard_D2s_v3`.
      * **Authentication type**: Select **Password**.
      * **Username**: Choose a username, for example, `azureadmin`.
      * **Password**: Create and confirm a strong password.
  * **Networking Tab**:
      * Ensure that for **NIC network security group**, "Advanced" is selected.
      * Under **Configure network security group**, ensure an inbound rule for **port 22 (SSH)** is allowed so you can connect.
  * Click **Review + create**, then **Create**.

#### 2\. Connect to Your VM

Once the VM is deployed, find its **Public IP address** on the overview page.

```bash
ssh your_username@YOUR_VM_PUBLIC_IP
```

Enter the password you created when prompted.

#### 3\. Install All Tools via Script

Create and run a single script to install everything we need.

1.  Create the script file: `nano setup_tools.sh`
2.  Paste the following code:
    ```bash
    #!/bin/bash
    # Update, install tools, and configure permissions
    sudo apt-get update && sudo apt-get upgrade -y
    sudo apt-get install -y git curl apt-transport-https docker.io
    wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    source ~/.profile
    rm go1.22.5.linux-amd64.tar.gz
    sudo usermod -aG docker ${USER}
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    rm kubectl
    curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
    sudo install minikube /usr/local/bin/minikube
    rm minikube
    echo "✅ Setup Complete! IMPORTANT: Please log out and log back in."
    ```
3.  Save and exit (`Ctrl+X`, `Y`, `Enter`).
4.  Make it executable: `chmod +x setup_tools.sh`
5.  Run it: `./setup_tools.sh`
6.  **CRITICAL**: After the script finishes, log out (`exit`) and log back in for the Docker permissions to apply.

🚨 **Troubleshooting: Docker Permissions**

  * **Problem:** You run a `docker` command (like `docker ps` or `minikube start`) and see an error like: `permission denied while trying to connect to the Docker daemon socket`.
  * **Cause:** Your current terminal session doesn't have the new group permissions.
  * **Solution:** The best fix is to log out and log back in. For a temporary fix in your current session, run `newgrp docker`.

-----

### Stage 2: Application Code & Container

1.  **Create Project Files**:

    ```bash
    mkdir ~/quote-api && cd ~/quote-api
    nano go.mod
    ```

    Paste the following (replace `your-github-username`):

    ```go
    module github.com/your-github-username/quote-api

    go 1.22
    ```

    Save and exit. Now create the main app file:

    ```bash
    nano main.go
    ```

    Paste the application code:

    ```go
    package main

    import (
        "fmt"
        "math/rand"
        "net/http"
        "time"
    )

    func quoteHandler(w http.ResponseWriter, r *http.Request) {
        quotes := []string{
            "The only way to do great work is to love what you do. - Steve Jobs",
            "Success is not final, failure is not fatal: it is the courage to continue that counts. - Winston Churchill",
        }
        rand.Seed(time.Now().UnixNano())
        randomQuote := quotes[rand.Intn(len(quotes))]
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"quote": "%s"}`, randomQuote)
    }

    func main() {
        http.HandleFunc("/", quoteHandler)
        fmt.Println("Starting Quote API server on port 8080...")
        http.ListenAndServe(":8080", nil)
    }
    ```

    Save and exit.

2.  **Create the Dockerfile**:
    This file is the recipe for building your application container.

    ```bash
    nano Dockerfile
    ```

    Paste the multi-stage build configuration:

    ```dockerfile
    # ---- Stage 1: The Builder ----
    FROM golang:1.22-alpine AS builder
    WORKDIR /app
    COPY go.mod ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

    # ---- Stage 2: The Final Image ----
    FROM scratch
    WORKDIR /app
    COPY --from=builder /app/server .
    EXPOSE 8080
    ENTRYPOINT ["/app/server"]
    ```

    Save and exit.

-----

### Stage 3: Version Control & CI Pipeline

1.  **Initialize Git and Create GitHub Repo**:

      * Inside your `~/quote-api` directory, run:
        ```bash
        git init -b main
        git config --global user.name "Your Name"
        git config --global user.email "your.email@example.com"
        ```
      * Go to GitHub and create a new, **empty** public repository named `quote-api`.

2.  **Create the CI Workflow**:

    ```bash
    mkdir -p .github/workflows
    nano .github/workflows/ci.yml
    ```

    Paste the following secure workflow configuration:

    ```yaml
    name: Build and Push Docker Image
    on:
      push:
        branches: [ "main" ]
    jobs:
      build-and-push:
        runs-on: ubuntu-latest
        steps:
          - name: Check out code
            uses: actions/checkout@v4
          - name: Log in to Docker Hub
            uses: docker/login-action@v3
            with:
              username: ${{ secrets.DOCKERHUB_USERNAME }}
              password: ${{ secrets.DOCKERHUB_TOKEN }}
          - name: Build and push Docker image
            uses: docker/build-push-action@v5
            with:
              context: .
              push: true
              tags: your-dockerhub-username/quote-api:latest # IMPORTANT: Change this!
    ```

    **Crucially, change `your-dockerhub-username` to your actual Docker Hub username.** Save and exit.

3.  **Configure Secrets**:

      * Go to your new GitHub repo \> **Settings** \> **Secrets and variables** \> **Actions**.
      * Create **New repository secret** `DOCKERHUB_USERNAME` with your Docker Hub username.
      * Create **New repository secret** `DOCKERHUB_TOKEN` with a Docker Hub Access Token.

4.  **Push Your Code**:

      * Link your local repo to GitHub and push:
        ```bash
        git remote add origin https://github.com/your-github-username/quote-api.git
        git add .
        git commit -m "Initial project setup"
        git push -u origin main
        ```

🚨 **Troubleshooting: Git Push**

  * **Problem:** `fatal: The current branch main has no upstream branch`.

  * **Cause:** Your local Git doesn't know where to push the `main` branch on the remote.

  * **Solution:** Use the command Git suggests: `git push --set-upstream origin main`. You only need to do this once.

  * **Problem:** GitHub prompts for a password and rejects it.

  * **Cause:** You must use a **Personal Access Token (PAT)** for command-line authentication, not your account password.

  * **Solution:** Go to GitHub \> Settings \> Developer settings \> Personal access tokens \> Generate new token. Give it the `repo` scope. Copy the token and paste it at the password prompt.

  * **Problem:** `remote rejected... Push cannot contain secrets`.

  * **Cause:** GitHub's Push Protection has detected a secret in your code. You tried to hardcode a password instead of using GitHub Secrets.

  * **Solution:** Follow step 3 above to create secrets. Ensure your `ci.yml` file uses `${{ secrets.YOUR_SECRET }}` and does not contain any real passwords or tokens. Use `git reset --soft HEAD~1` to undo the bad commit, fix the file, and commit again.

-----

### Stage 4: Kubernetes & ArgoCD Setup

1.  **Start Minikube**:
    ```bash
    minikube start --memory=4096 --cpus=2
    ```

🚨 **Troubleshooting: Minikube Start**

  * **Problem:** Minikube shows `STATUS: Pending` for pods.
  * **Cause:** The cluster doesn't have enough CPU or Memory.
  * **Solution:** Delete the cluster with `minikube delete` and restart it, allocating more resources as shown in the command above.

<!-- end list -->

2.  **Install ArgoCD**:

    ```bash
    kubectl create namespace argocd
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
    ```

3.  **Create Kubernetes Manifests**:
    These files tell Kubernetes how to run your application.

    ```bash
    cd ~/quote-api
    mkdir k8s
    nano k8s/deployment.yaml
    ```

    Paste the deployment configuration (change `your-dockerhub-username`):

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: quote-api-deployment
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: quote-api
      template:
        metadata:
          labels:
            app: quote-api
        spec:
          containers:
          - name: quote-api-container
            image: your-dockerhub-username/quote-api:latest # IMPORTANT: Change this!
            ports:
            - containerPort: 8080
    ```

    Save and exit. Now create the service:

    ```bash
    nano k8s/service.yaml
    ```

    Paste the service configuration:

    ```yaml
    apiVersion: v1
    kind: Service
    metadata:
      name: quote-api-service
    spec:
      selector:
        app: quote-api
      ports:
        - protocol: TCP
          port: 80
          targetPort: 8080
      type: NodePort
    ```

    Save and exit.

4.  **Push Manifests to GitHub**:

    ```bash
    git add k8s/
    git commit -m "Add Kubernetes manifests"
    git push
    ```

-----

### Stage 5: GitOps Deployment & Final Loop

1.  **Access ArgoCD UI**:

      * Open firewall port `8080` in your Azure VM's **Networking** settings.
      * Run the port-forward command:
        ```bash
        kubectl port-forward svc/argocd-server -n argocd 8080:443 --address='0.0.0.0' &
        ```
      * Get the password:
        ```bash
        kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
        ```
      * Log in at `http://YOUR_VM_PUBLIC_IP:8080` with username `admin` and the password.

2.  **Create ArgoCD Application**:

      * In the UI, click **+ NEW APP**.
      * **Application Name**: `quote-api`
      * **Project Name**: `default`
      * **Sync Policy**: `Automatic` (check `Prune Resources` and `Self Heal`).
      * **Repository URL**: `https://github.com/your-github-username/quote-api.git`
      * **Path**: `k8s`
      * **Cluster URL**: `https://kubernetes.default.svc`
      * **Namespace**: `default`
      * Click **CREATE**. The app will sync and become `Healthy`.

3.  **Access Your Application**:

🚨 **Troubleshooting: Accessing the Service**

  * **Problem:** Browser shows `connection has timed out`.
  * **Cause:** This is almost always a firewall or networking issue.
  * **Solution:**
    1.  **Run Minikube Tunnel**: In a **new, separate terminal**, run `minikube tunnel`. Leave it running. This is the most reliable way to expose services on a cloud VM.
    2.  **Open the Firewall Port**: Get the service port by running `minikube service quote-api-service --url`. It will give a URL like `http://192.168.49.2:31080`. The port is `31080`. Go to your Azure VM's **Networking** settings and create a new inbound rule to allow **TCP** traffic on this port (e.g., `31080`).
    3.  **Use the Right URL**: Combine your VM's **Public IP** with the port: `http://YOUR_VM_PUBLIC_IP:31080`.

<!-- end list -->

4.  **Test the Full Loop**:
      * Change the `main.go` file to add a new quote.
      * Run `git add .`, `git commit -m "New quote"`, and `git push`.
      * Watch GitHub Actions build a new image.
      * Watch ArgoCD automatically deploy it.
      * Refresh your browser to see the new quote\!
