#include "mainwindow.h"
#include "libraryview.h"
#include "playerwidget.h"
#include "connectiondialog.h"
#include "apiclient.h"
#include "settings.h"

#include <QMenuBar>
#include <QAction>
#include <QMessageBox>
#include <QCloseEvent>
#include <QApplication>
#include <QStyle>
#include <QPainter>
#include <QPixmap>
#include <QFont>
#include <QDebug>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , m_centralStack(new QStackedWidget(this))
    , m_libraryView(nullptr)
    , m_playerWidget(nullptr)
    , m_apiClient(new ApiClient(this))
    , m_settings(new Settings(this))
    , m_statusLabel(new QLabel("Not connected"))
    , m_trayIcon(new QSystemTrayIcon(this))
    , m_trayMenu(new QMenu(this))
{
    setupUi();
    setupToolbar();
    setupTrayIcon();
    connectSignals();

    resize(1200, 800);
    setWindowTitle("ber-client");
    setWindowIcon(createTrayIcon());

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

QIcon MainWindow::createTrayIcon() {
    return QIcon(":/ber.ico");
}

void MainWindow::setupUi() {
    m_libraryView = new LibraryView(m_apiClient, this);
    m_playerWidget = new PlayerWidget(m_apiClient, this);

    m_centralStack->addWidget(m_libraryView);
    m_centralStack->addWidget(m_playerWidget);
    m_centralStack->setCurrentWidget(m_libraryView);

    setCentralWidget(m_centralStack);
    statusBar()->addPermanentWidget(m_statusLabel);
}

void MainWindow::setupToolbar() {
    m_toolbar = addToolBar("Main");
    m_toolbar->setMovable(false);

    m_toolbar->addAction("Connect", this, &MainWindow::showConnectionDialog);
    m_toolbar->addSeparator();
    m_toolbar->addAction("Library", this, [this]() {
        m_centralStack->setCurrentWidget(m_libraryView);
    });
}

void MainWindow::setupTrayIcon() {
    QIcon icon = createTrayIcon();
    m_trayIcon->setIcon(icon);
    m_trayIcon->setToolTip("ber-client");

    QAction *showAction = m_trayMenu->addAction(style()->standardIcon(QStyle::SP_ComputerIcon), "Show ber");
    m_trayMenu->addSeparator();
    QAction *libraryAction = m_trayMenu->addAction(style()->standardIcon(QStyle::SP_DirIcon), "Manage Library");
    m_trayMenu->addSeparator();
    QAction *aboutAction = m_trayMenu->addAction(style()->standardIcon(QStyle::SP_MessageBoxInformation), "About");
    QAction *quitAction = m_trayMenu->addAction(style()->standardIcon(QStyle::SP_DialogCloseButton), "Quit");

    m_trayIcon->setContextMenu(m_trayMenu);

    connect(showAction, &QAction::triggered, this, &MainWindow::toggleWindowVisibility);
    connect(libraryAction, &QAction::triggered, this, [this]() {
        showNormal();
        activateWindow();
        raise();
        m_centralStack->setCurrentWidget(m_libraryView);
    });
    connect(aboutAction, &QAction::triggered, this, &MainWindow::showAboutDialog);
    connect(quitAction, &QAction::triggered, qApp, &QApplication::quit);

    connect(m_trayIcon, &QSystemTrayIcon::activated, this, [this](QSystemTrayIcon::ActivationReason reason) {
        if (reason == QSystemTrayIcon::DoubleClick) {
            toggleWindowVisibility();
        }
    });

    if (!QSystemTrayIcon::isSystemTrayAvailable()) {
        qWarning() << "System tray is not available on this system";
        return;
    }

    m_trayIcon->show();
    m_trayIcon->showMessage("ber-client", "Running in system tray", QSystemTrayIcon::Information, 3000);
}

void MainWindow::connectSignals() {
    connect(m_libraryView, &LibraryView::videoSelected, this, [this](const QString &videoId) {
        m_playerWidget->playVideo(videoId);
        m_centralStack->setCurrentWidget(m_playerWidget);
    });

    connect(m_playerWidget, &PlayerWidget::backToLibrary, this, [this]() {
        m_centralStack->setCurrentWidget(m_libraryView);
    });

    connect(m_apiClient, &ApiClient::connected, this, &MainWindow::onConnected);
    connect(m_apiClient, &ApiClient::disconnected, this, &MainWindow::onDisconnected);
    connect(m_apiClient, &ApiClient::connectionFailed, this, [this](const QString &error) {
        m_statusLabel->setText("Connection failed: " + error);
    });

    connect(m_playerWidget, &PlayerWidget::fullscreenToggled, this, &MainWindow::onFullscreenToggled);
}

void MainWindow::closeEvent(QCloseEvent *event) {
    if (m_trayIcon->isVisible()) {
        hide();
        event->ignore();
    } else {
        event->accept();
    }
}

void MainWindow::toggleWindowVisibility() {
    if (isVisible()) {
        hide();
    } else {
        showNormal();
        activateWindow();
        raise();
    }
}

void MainWindow::showConnectionDialog() {
    ConnectionDialog dialog(m_apiClient, m_settings, this);
    dialog.exec();
}

void MainWindow::onConnected(const QString &serverUrl) {
    m_serverUrl = serverUrl;
    m_statusLabel->setText("Connected: " + serverUrl);
    m_settings->setLastServer(serverUrl);
    m_libraryView->refresh();
}

void MainWindow::onDisconnected() {
    m_statusLabel->setText("Disconnected");
    m_centralStack->setCurrentWidget(m_libraryView);
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
