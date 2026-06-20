#pragma once

#include <QMainWindow>
#include <QStackedWidget>
#include <QToolBar>
#include <QStatusBar>
#include <QLabel>

class LibraryView;
class PlayerWidget;
class ConnectionDialog;
class ApiClient;
class Settings;

class MainWindow : public QMainWindow {
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);

protected:
    void closeEvent(QCloseEvent *event) override;

private slots:
    void showConnectionDialog();
    void onConnected(const QString &serverUrl);
    void onDisconnected();
    void onFullscreenToggled(bool fullscreen);
    void showAboutDialog();

private:
    void setupUi();
    void setupToolbar();
    void connectSignals();

    QStackedWidget *m_centralStack;
    LibraryView *m_libraryView;
    PlayerWidget *m_playerWidget;
    ApiClient *m_apiClient;
    Settings *m_settings;
    QLabel *m_statusLabel;
    QToolBar *m_toolbar;
    QString m_serverUrl;
};
