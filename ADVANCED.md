# Advanced Usage

Enhance your Bangs experience by simplifying access and ensuring the service runs seamlessly in the background.
This guide provides methods to map a memorable hostname to your local Bangs server and ensures it starts automatically on system boot.

## 1. Simplify Access with Host Entries

Instead of typing `localhost:8080` or `s.dikka.dev`, assign a custom, easy-to-remember hostname to your local Bangs server. Follow the instructions for your operating system below:

### **macOS**

> **Note:** I have no experience with macOS configurations, so please take these commands with a grain of salt. Always verify and understand commands before executing them.

1. **Edit the Hosts File**

    Open the Terminal and run:

    ```bash
    sudo vim /etc/hosts
    ```

2. **Add Host Entry**

    Add the following line at the end of the file:

    ```
    127.0.0.1   bangs
    ```

3. **Save and Exit**

    - In `vim`, press `Esc`, type `:x`, and press `Enter` to save and exit.

4. **Flush DNS Cache**

    ```bash
    sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder
    ```

### **Linux**

1. **Edit the Hosts File**

    Open the Terminal and run:

    ```bash
    sudo vim /etc/hosts
    ```

2. **Add Host Entry**

    Add the following line at the end of the file:

    ```
    127.0.0.1   bangs
    ```

3. **Save and Exit**

    - In `vim`, press `Esc`, type `:x`, and press `Enter` to save and exit.

4. **Flush DNS Cache**

    The method varies based on your distribution. For example, on Ubuntu:

    ```bash
    sudo systemd-resolve --flush-caches
    ```

### **Windows 11**

1. **Edit the Hosts File**

    - Open **Notepad** as an administrator.
    - Open the file located at `C:\Windows\System32\drivers\etc\hosts`.

2. **Add Host Entry**

    Add the following line at the end of the file:

    ```
    127.0.0.1   bangs
    ```

3. **Save and Exit**

    - Save the file and close Notepad.

4. **Flush DNS Cache**

    Open **Command Prompt** as an administrator and run:

    ```cmd
    ipconfig /flushdns
    ```

### **Usage Example**

After setting up the host entry, you can access your local Bangs server using:


```
http://bangs/?q=!gh Sett17/bangs'
```

## 2. Run Bangs as a Persistent Service

Ensure Bangs starts automatically and runs in the background using the following methods based on your operating system:

> **Note**: These _should_ work, but I have not tried all of the mmyself. If you see any errors in these setups, please open an issue.

### **Using Docker (All Platforms)**

Run Bangs in a Docker container with a restart policy to ensure it stays running:

```bash
docker run -d \
  --name bangs \
  -p 8080:8080 \
  -v /path/to/your/bangs.yaml:/app/bangs.yaml \
  --restart unless-stopped \
  TBD:latest
```
> **Note**: Replace /path/to/your/bangs.yaml with the actual path to your bangs.yaml file.

### **macOS**

> **Note:** I have no experience with macOS configurations, so please take these commands with a grain of salt. Always verify and understand commands before executing them.

Use `launchd` to create a persistent service:

1. **Create a Launch Agent File**

    ```bash
    vim ~/Library/LaunchAgents/com.sett17.bangs.plist
    ```

2. **Add Configuration**

    Paste the following content into the file:

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
    <plist version="1.0">
      <dict>
        <key>Label</key>
        <string>com.sett17.bangs</string>
        <key>ProgramArguments</key>
        <array>
          <string>/path/to/bangs</string>
          <string>--bangs</string>
          <string>/path/to/bangs.yaml</string>
          <string>--port</string>
          <string>8080</string>
        </array>
        <key>RunAtLoad</key>
        <true/>
        <key>KeepAlive</key>
        <true/>
        <key>StandardOutPath</key>
        <string>/tmp/bangs.log</string>
        <key>StandardErrorPath</key>
        <string>/tmp/bangs.err</string>
      </dict>
    </plist>
    ```

    > **Note:** Replace `/path/to/bangs` and `/path/to/bangs.yaml` with the actual paths.

3. **Load the Launch Agent**

    ```bash
    launchctl load ~/Library/LaunchAgents/com.sett17.bangs.plist
    ```

### **Linux**

Use `systemd` to create a persistent service:

1. **Create a Service File**

    ```bash
    sudo vim /etc/systemd/system/bangs.service
    ```

2. **Add Configuration**

    Paste the following content into the file:

    ```ini
    [Unit]
    Description=Bangs Search Engine
    After=network.target

    [Service]
    ExecStart=/path/to/bangs --bangs /path/to/bangs.yaml --port 8080
    Restart=always
    User=yourusername
    Environment=PATH=/usr/bin:/usr/local/bin
    WorkingDirectory=/path/to/working/directory

    [Install]
    WantedBy=multi-user.target
    ```

    > **Note:** Replace `/path/to/bangs`, `/path/to/bangs.yaml`, and `/path/to/working/directory` with the actual paths. Replace `yourusername` with your actual username.

3. **Enable and Start the Service**

    ```bash
    sudo systemctl enable bangs
    sudo systemctl start bangs
    ```

### **Windows 11**

Use the **Windows Task Scheduler** to create a persistent service:

1. **Open Task Scheduler**

    Press `Win + R`, type `taskschd.msc`, and press `Enter`.

2. **Create a New Task**

    - Click on **"Create Task..."** in the **Actions** pane.

3. **Configure General Settings**

    - **Name**: Bangs Search Engine
    - **Description**: Runs Bangs search engine as a background service.
    - **Security options**: Select **"Run whether user is logged on or not"**.
    - **Check**: **"Run with highest privileges"**.

4. **Configure Triggers**

    - Go to the **"Triggers"** tab.
    - Click **"New..."**.
    - **Begin the task**: At startup.
    - Click **"OK"**.

5. **Configure Actions**

    - Go to the **"Actions"** tab.
    - Click **"New..."**.
    - **Action**: Start a program.
    - **Program/script**: `C:\Path\To\bangs.exe`
    - **Add arguments**: `--bangs C:\Path\To\bangs.yaml --port 8080`
    - Click **"OK"**.

6. **Configure Conditions and Settings**

    - Adjust any additional settings as needed, such as restarting the task on failure.

7. **Save the Task**

    - Click **"OK"**.
    - Enter your password if prompted.

8. **Start the Task**

    - In Task Scheduler, locate your newly created task.
    - Right-click and select **"Run"** to start it immediately.

## 3. Ensure Automatic Startup

By setting up Bangs as a persistent service using Docker, `launchd`, `systemd`, or Task Scheduler, you ensure that Bangs starts automatically when your system boots, providing uninterrupted search capabilities.

---

By following these advanced usage steps, you can streamline your Bangs setup, making it more accessible and reliable across different platforms.
