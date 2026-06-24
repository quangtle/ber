#include "mainwindow.h"
#include "libraryview.h"
#include "playerbar.h"
#include "playerwidget.h"
#include "connectiondialog.h"
#include "apiclient.h"
#include "settings.h"

#include <QAction>
#include <QVBoxLayout>
#include <QMessageBox>
#include <QCloseEvent>
#include <QApplication>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , m_centralStack(new QStackedWidget(this))
    , m_libraryPage(new QWidget(this))
    , m_libraryView(nullptr)
    , m_playerBar(nullptr)
    , m_fullPlayer(nullptr)
    , m_apiClient(new ApiClient(this))
    , m_settings(new Settings(this))
    , m_statusLabel(new QLabel("Not connected"))
{
    setupUi();
    setupToolbar();
    connectSignals();

    resize(1200, 800);
    setWindowTitle("ber-client");

    // Auto-connect: try saved server, then discover, then give up
    m_statusLabel->setText("Connecting...");
    QString server;
    QString saved = m_settings->lastServer();

    if (!saved.isEmpty() && m_apiClient->tryConnect(saved)) {
        server = saved;
    } else {
        server = m_apiClient->discoverServer();
    }

    if (!server.isEmpty()) {
        m_apiClient->connectToServer(server);
    } else {
        m_statusLabel->setText("No server found — click Connect to enter address");
    }
}

void MainWindow::setupUi() {
    // Library page: thumbnail grid on top, player bar on bottom
    m_libraryView = new LibraryView(m_apiClient, m_libraryPage);
    m_playerBar = new PlayerBar(m_apiClient, m_libraryPage);
    m_playerBar->hide();

    auto *libLayout = new QVBoxLayout(m_libraryPage);
    libLayout->setContentsMargins(0, 0, 0, 0);
    libLayout->setSpacing(0);
    libLayout->addWidget(m_libraryView, 1);
    libLayout->addWidget(m_playerBar);

    m_centralStack->addWidget(m_libraryPage);  // page 0
    m_fullPlayer = new PlayerWidget(m_apiClient, this);
    m_centralStack->addWidget(m_fullPlayer);   // page 1
    m_centralStack->setCurrentWidget(m_libraryPage);

    setCentralWidget(m_centralStack);
    statusBar()->addPermanentWidget(m_statusLabel);
}

void MainWindow::setupToolbar() {
    m_toolbar = addToolBar("Main");
    m_toolbar->setMovable(false);

    m_toolbar->addAction("Connect", this, &MainWindow::showConnectionDialog);
}

void MainWindow::connectSignals() {
    // Click a video → fullscreen player
    connect(m_libraryView, &LibraryView::videoSelected, this, [this](const QString &videoId) {
        m_lastVideoId = videoId;
        m_fullPlayer->playVideo(videoId);
        m_centralStack->setCurrentWidget(m_fullPlayer);
    });

    // "← Library" in fullscreen → back to library, play in bottom bar
    connect(m_fullPlayer, &PlayerWidget::backToLibrary, this, [this]() {
        qint64 pos = m_fullPlayer->position();
        m_centralStack->setCurrentWidget(m_libraryPage);
        m_playerBar->playVideo(m_lastVideoId, pos);
        m_playerBar->show();
        m_fullPlayer->stop();
    });

    // Close button on bottom bar → stop and hide
    connect(m_playerBar, &PlayerBar::closeClicked, this, [this]() {
        m_playerBar->stop();
        m_playerBar->hide();
    });

    connect(m_apiClient, &ApiClient::connected, this, &MainWindow::onConnected);
    connect(m_apiClient, &ApiClient::disconnected, this, &MainWindow::onDisconnected);
    connect(m_apiClient, &ApiClient::connectionFailed, this, [this](const QString &error) {
        m_statusLabel->setText("Connection failed: " + error);
    });

    connect(m_fullPlayer, &PlayerWidget::fullscreenToggled, this, &MainWindow::onFullscreenToggled);
}

void MainWindow::closeEvent(QCloseEvent *event) {
    QApplication::quit();
    event->accept();
}

void MainWindow::showConnectionDialog() {
    ConnectionDialog dialog(m_apiClient, m_settings, this);
    dialog.exec();
}

void MainWindow::onConnected(const QString &serverUrl) {
    m_statusLabel->setText("Connected: " + serverUrl);
    m_settings->setLastServer(serverUrl);
    m_libraryView->refresh();
}

void MainWindow::onDisconnected() {
    m_statusLabel->setText("Disconnected");
    m_centralStack->setCurrentWidget(m_libraryPage);
}

void MainWindow::onFullscreenToggled(bool fullscreen) {
    m_toolbar->setVisible(!fullscreen);
    statusBar()->setVisible(!fullscreen);
}

void MainWindow::showAboutDialog() {
    QMessageBox::about(this, "About ber-client",
        "ber-client " + QApplication::applicationVersion() + "\n\n"
        "A lightweight local streaming client.\n"
        "Connect to a ber-server to browse and play videos.");
}
