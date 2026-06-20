#pragma once

#include <QObject>
#include <QString>
#include <QJsonObject>
#include <QList>
#include <QNetworkAccessManager>
#include <QUdpSocket>
#include <functional>

class ApiClient : public QObject {
    Q_OBJECT

public:
    explicit ApiClient(QObject *parent = nullptr);

    void connectToServer(const QString &address);
    bool isConnected() const;
    bool tryConnect(const QString &address);
    QString discoverServer();

    void fetchLibrary();
    void fetchVideoInfo(const QString &videoId, std::function<void(const QJsonObject &)> callback);
    QString streamUrl(const QString &videoId) const;

signals:
    void connected(const QString &serverUrl);
    void disconnected();
    void libraryLoaded(const QList<QJsonObject> &videos);
    void connectionFailed(const QString &error);

private:
    QNetworkAccessManager *m_networkManager;
    QString m_serverUrl;
    bool m_connected;
};
