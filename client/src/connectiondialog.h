#pragma once

#include <QDialog>
#include <QLineEdit>
#include <QPushButton>
#include <QLabel>

class ApiClient;
class Settings;

class ConnectionDialog : public QDialog {
    Q_OBJECT

public:
    explicit ConnectionDialog(ApiClient *client, Settings *settings, QWidget *parent = nullptr);

private slots:
    void onConnect();
    void onAutoDetect();

private:
    void setupUi();

    ApiClient *m_apiClient;
    Settings *m_settings;
    QLineEdit *m_serverEdit;
    QPushButton *m_connectBtn;
    QPushButton *m_autoDetectBtn;
    QLabel *m_statusLabel;
};
