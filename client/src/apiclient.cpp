#include "apiclient.h"

#include <QNetworkReply>
#include <QJsonDocument>
#include <QJsonArray>
#include <QJsonObject>
#include <QUrl>
#include <QHostAddress>
#include <QEventLoop>
#include <QTimer>

ApiClient::ApiClient(QObject *parent)
    : QObject(parent)
    , m_networkManager(new QNetworkAccessManager(this))
    , m_connected(false)
{
}

void ApiClient::connectToServer(const QString &address) {
    // ponytail: normalize address, may have http:// prefix or not
    m_serverUrl = address.startsWith("http://") ? address : "http://" + address;

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

bool ApiClient::tryConnect(const QString &address) {
    QString base = address.startsWith("http://") ? address : "http://" + address;
    QString url = base + "/api/status";
    QUrl qurl(url);
    QNetworkRequest request(qurl);
    QNetworkReply *reply = m_networkManager->get(request);

    QEventLoop loop;
    QTimer timer;
    timer.setSingleShot(true);
    connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    connect(&timer, &QTimer::timeout, &loop, &QEventLoop::quit);
    timer.start(2000);

    loop.exec();

    bool ok = reply->isFinished() && reply->error() == QNetworkReply::NoError;
    reply->deleteLater();
    return ok;
}

QString ApiClient::discoverServer() {
    QUdpSocket socket;
    if (!socket.bind(QHostAddress::AnyIPv4, 10001, QUdpSocket::ShareAddress)) {
        return QString();
    }

    QByteArray buffer(256, '\0');
    for (int i = 0; i < 6; i++) {
        if (!socket.waitForReadyRead(500)) {
            continue;
        }
        while (socket.hasPendingDatagrams()) {
            QHostAddress sender;
            quint16 senderPort;
            qint64 len = socket.readDatagram(buffer.data(), buffer.size(), &sender, &senderPort);
            QString msg = QString::fromUtf8(buffer.left(len));
            if (msg.startsWith("ber-server:")) {
                // beacon payload is the listen addr (e.g. ":8080"), combine with sender IP
                return sender.toString() + msg.mid(11);
            }
        }
    }
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

QString ApiClient::thumbnailUrl(const QString &videoId) const {
    return m_serverUrl + "/api/stream/" + videoId + "/thumbnail";
}
