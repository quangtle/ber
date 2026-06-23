#include "libraryview.h"
#include "apiclient.h"

#include <QVBoxLayout>
#include <QLabel>
#include <QFont>
#include <QJsonArray>
#include <QJsonObject>
#include <QNetworkReply>
#include <QPixmap>
#include <QIcon>
#include <QListView>

LibraryView::LibraryView(ApiClient *client, QWidget *parent)
    : QWidget(parent)
    , m_apiClient(client)
    , m_gridWidget(new QListWidget(this))
    , m_nam(new QNetworkAccessManager(this))
{
    setupUi();
    connect(m_apiClient, &ApiClient::libraryLoaded, this, &LibraryView::onLibraryLoaded);
    connect(m_gridWidget, &QListWidget::itemClicked, this, &LibraryView::onItemClicked);
}

void LibraryView::setupUi() {
    auto *layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);

    auto *title = new QLabel("Video Library");
    QFont f = title->font();
    f.setPointSize(16);
    f.setBold(true);
    title->setFont(f);
    title->setContentsMargins(16, 16, 16, 8);
    layout->addWidget(title);

    m_gridWidget->setViewMode(QListView::IconMode);
    m_gridWidget->setIconSize(QSize(178, 100));
    m_gridWidget->setGridSize(QSize(200, 140));
    m_gridWidget->setMovement(QListView::Static);
    m_gridWidget->setResizeMode(QListView::Adjust);
    m_gridWidget->setWordWrap(true);
    m_gridWidget->setSpacing(0);
    m_gridWidget->setWrapping(true);
    m_gridWidget->setFlow(QListView::LeftToRight);

    layout->addWidget(m_gridWidget);
}

void LibraryView::refresh() {
    m_apiClient->fetchLibrary();
}

void LibraryView::onLibraryLoaded(const QList<QJsonObject> &videos) {
    m_gridWidget->clear();
    for (const auto &v : videos) {
        auto *item = new QListWidgetItem(v["title"].toString());
        item->setData(Qt::UserRole, v["id"].toString());
        item->setData(Qt::UserRole + 1, v["file_path"].toString());
        m_gridWidget->addItem(item);

        loadThumbnail(v["id"].toString(), item);
    }
}

void LibraryView::loadThumbnail(const QString &videoId, QListWidgetItem *item) {
    QNetworkRequest req(QUrl(m_apiClient->thumbnailUrl(videoId)));
    QNetworkReply *reply = m_nam->get(req);
    connect(reply, &QNetworkReply::finished, this, [reply, item]() {
        if (reply->error() == QNetworkReply::NoError) {
            QPixmap pix;
            if (pix.loadFromData(reply->readAll(), "JPEG")) {
                item->setIcon(QIcon(pix));
            }
        }
        reply->deleteLater();
    });
}

void LibraryView::onItemClicked(QListWidgetItem *item) {
    if (!item) return;
    QString videoId = item->data(Qt::UserRole).toString();
    emit videoSelected(videoId);
}
