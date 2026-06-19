#include "connectiondialog.h"
#include "apiclient.h"
#include "settings.h"

#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QFormLayout>

ConnectionDialog::ConnectionDialog(ApiClient *client, Settings *settings, QWidget *parent)
    : QDialog(parent)
    , m_apiClient(client)
    , m_settings(settings)
{
    setupUi();
    QString saved = m_settings->lastServer();
    if (!saved.isEmpty()) {
        m_serverEdit->setText(saved);
    }
}

void ConnectionDialog::setupUi() {
    setWindowTitle("Connect to ber-server");
    setFixedSize(400, 200);

    auto *layout = new QVBoxLayout(this);

    auto *form = new QFormLayout();
    m_serverEdit = new QLineEdit("localhost:8080");
    form->addRow("Server Address:", m_serverEdit);
    layout->addLayout(form);

    m_statusLabel = new QLabel("Enter a server address or use auto-detect");
    m_statusLabel->setWordWrap(true);
    layout->addWidget(m_statusLabel);

    auto *btnLayout = new QHBoxLayout();
    m_autoDetectBtn = new QPushButton("Auto-Detect");
    m_connectBtn = new QPushButton("Connect");
    btnLayout->addWidget(m_autoDetectBtn);
    btnLayout->addStretch();
    btnLayout->addWidget(m_connectBtn);
    layout->addLayout(btnLayout);

    connect(m_connectBtn, &QPushButton::clicked, this, &ConnectionDialog::onConnect);
    connect(m_autoDetectBtn, &QPushButton::clicked, this, &ConnectionDialog::onAutoDetect);
}

void ConnectionDialog::onConnect() {
    QString addr = m_serverEdit->text().trimmed();
    if (addr.isEmpty()) return;

    m_statusLabel->setText("Connecting...");
    m_connectBtn->setEnabled(false);

    m_apiClient->connectToServer(addr);
    if (m_apiClient->isConnected()) {
        accept();
    } else {
        m_statusLabel->setText("Connection failed. Check the address and try again.");
        m_connectBtn->setEnabled(true);
    }
}

void ConnectionDialog::onAutoDetect() {
    m_statusLabel->setText("Scanning for servers on the local network...");
    m_autoDetectBtn->setEnabled(false);

    QString found = m_apiClient->discoverServer();
    if (!found.isEmpty()) {
        m_serverEdit->setText(found);
        m_statusLabel->setText("Found server at " + found);
    } else {
        m_statusLabel->setText("No servers found. Enter address manually.");
    }
    m_autoDetectBtn->setEnabled(true);
}
