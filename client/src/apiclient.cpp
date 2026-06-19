#include "apiclient.h"

#include <QNetworkReply>
#include <QJsonDocument>
#include <QJsonArray>
#include <QJsonObject>
#include <QUrl>

ApiClient::ApiClient(QObject *parent)
    : QObject(parent)
    , m_networkManager(new QNetworkAccessManager(this))
    , m_connected(false)
{
}

void ApiClient::connectToServer(const QString &address) {
    m_serverUrl = "http://" + address;

    QNetworkRequest request(QUrl(m_serverUrl + "/api/status"));
    QNetworkReply *reply = m_networkManager->get(request);

    connect(reply, &QNetworkReply::finished, this, [this, reply, address]() {
        if (reply->error() == QNetworkReply::NoError) {
            m_connected = true;
            emit connected(m_serverUrl);
        } else {
            m_connected = false;
            emit connectionFailed(reply->errorString());
        }
        reply->deleteLater();
    });
}

bool ApiClient::isConnected() const {
    return m_connected;
}

QString ApiClient::discoverServer() {
    return QString();
}

void ApiClient::fetchLibrary() {
    if (!m_connected) return;

    QNetworkRequest request(QUrl(m_serverUrl + "/api/library"));
    QNetworkReply *reply = m_networkManager->get(request);

    connect(reply, &QNetworkReply::finished, this, [this, reply]() {
        if (reply->error() == QNetworkReply::NoError) {
            QJsonDocument doc = QJsonDocument::fromJson(reply->readAll());
            QJsonObject obj = doc.object();
            if (obj["ok"].toBool()) {
                QJsonArray arr = obj["data"].toArray();
                QList<QJsonObject> videos;
                for (const auto &v : arr) {
                    videos.append(v.toObject());
                }
                emit libraryLoaded(videos);
            }
        }
        reply->deleteLater();
    });
}

void ApiClient::fetchVideoInfo(const QString &videoId, std::function<void(const QJsonObject &)> callback) {
    if (!m_connected) return;

    QNetworkRequest request(QUrl(m_serverUrl + "/api/library/" + videoId));
    QNetworkReply *reply = m_networkManager->get(request);

    connect(reply, &QNetworkReply::finished, this, [reply, callback]() {
        if (reply->error() == QNetworkReply::NoError) {
            QJsonDocument doc = QJsonDocument::fromJson(reply->readAll());
            QJsonObject obj = doc.object();
            if (obj["ok"].toBool()) {
                callback(obj["data"].toObject());
            }
        }
        reply->deleteLater();
    });
}

QString ApiClient::streamUrl(const QString &videoId) const {
    return m_serverUrl + "/api/stream/" + videoId;
}
