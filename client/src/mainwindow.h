#pragma once

#include <QMainWindow>
#include <QStackedWidget>
#include <QToolBar>
#include <QStatusBar>
#include <QLabel>
#include <QString>

class LibraryView;
class PlayerBar;
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
    QWidget *m_libraryPage;
    LibraryView *m_libraryView;
    PlayerBar *m_playerBar;
    PlayerWidget *m_fullPlayer;
    ApiClient *m_apiClient;
    Settings *m_settings;
    QLabel *m_statusLabel;
    QToolBar *m_toolbar;
    QString m_lastVideoId;
};
