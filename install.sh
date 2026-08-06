#!/bin/sh

set -e

BASE_URL="https://wgri.de/packages"
PLATFORM="linux"
LATEST_VERSION=$(curl -s ${BASE_URL}/changelog/VERSION.txt)

if [ "$(uname -m)" = "x86_64" ]; then
    ARCH="amd64"
else
    ARCH="arm64"
fi


# DOWNLOAD
echo "Downloading watchgrid for ${PLATFORM}/${ARCH}..."
curl -sSfL "${BASE_URL}/${PLATFORM}/${ARCH}/watchgrid" -o "/tmp/watchgrid-${LATEST_VERSION}"

echo "Verifying checksum..."
EXPECTED_SHA=$(curl -sSfL "${BASE_URL}/${PLATFORM}/${ARCH}/watchgrid.sha256" | awk '{print $1}')
if [ -z "$EXPECTED_SHA" ]; then
    echo "Error: Could not retrieve checksum from server."
    rm -f "/tmp/watchgrid-${LATEST_VERSION}"
    exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_SHA=$(sha256sum "/tmp/watchgrid-${LATEST_VERSION}" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_SHA=$(shasum -a 256 "/tmp/watchgrid-${LATEST_VERSION}" | awk '{print $1}')
else
    echo "Error: Neither sha256sum nor shasum command found."
    rm -f "/tmp/watchgrid-${LATEST_VERSION}"
    exit 1
fi

if [ "${EXPECTED_SHA}" != "${ACTUAL_SHA}" ]; then
    echo "Error: Checksum verification failed!"
    echo "Expected: ${EXPECTED_SHA}"
    echo "Actual:   ${ACTUAL_SHA}"
    rm -f "/tmp/watchgrid-${LATEST_VERSION}"
    exit 1
fi

# REMOVE OLD VERSIONS & INSTALL
chmod +x /tmp/watchgrid-${LATEST_VERSION}
if [ -f /usr/local/bin/watchgrid ]; then
    sudo rm /usr/local/bin/watchgrid*
fi
sudo mv /tmp/watchgrid-${LATEST_VERSION} /usr/local/bin/watchgrid-${LATEST_VERSION}
sudo ln -sf /usr/local/bin/watchgrid-${LATEST_VERSION} /usr/local/bin/watchgrid


# CONFIG
mkdir -p /etc/watchgrid
if [ ! -f /etc/watchgrid/agent.json ]; then
sudo cat << EOF > /etc/watchgrid/agent.json
{
    "agent_id": "${WG_AGENT}",
    "token": "${WG_TOKEN}",
    "base_url": "https://wgri.de/api"
}
EOF
fi


# SYSTEMD SERVICE (DEPRECATED TIMER IN FAVOR OF SERVICE DAEMON)
if [ -f /etc/systemd/system/watchgrid.timer ]; then
    sudo systemctl stop watchgrid.timer || true
    sudo systemctl disable watchgrid.timer || true
    sudo rm -f /etc/systemd/system/watchgrid.timer
fi

sudo cat << EOF > /etc/systemd/system/watchgrid.service
[Unit]
Description=WatchGrid Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/watchgrid
Type=simple
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable watchgrid.service
sudo systemctl start watchgrid.service


echo "WatchGrid Agent with id ${AGENT_ID} in version ${LATEST_VERSION} installed successfully."