#include "libraryview.h"
#include "apiclient.h"

#include <QVBoxLayout>
#include <QLabel>
#include <QFont>
#include <QJsonArray>
#include <QJsonObject>

LibraryView::LibraryView(ApiClient *client, QWidget *parent)
    : QWidget(parent)
    , m_apiClient(client)
    , m_listWidget(new QListWidget(this))
{
    setupUi();
    connect(m_apiClient, &ApiClient::libraryLoaded, this, &LibraryView::onLibraryLoaded);
    connect(m_listWidget, &QListWidget::itemClicked, this, &LibraryView::onItemClicked);
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

    layout->addWidget(m_listWidget);
}

void LibraryView::refresh() {
    m_apiClient->fetchLibrary();
}

void LibraryView::onLibraryLoaded(const QList<QJsonObject> &videos) {
    m_listWidget->clear();
    for (const auto &v : videos) {
        auto *item = new QListWidgetItem(v["title"].toString());
        item->setData(Qt::UserRole, v["id"].toString());
        item->setData(Qt::UserRole + 1, v["file_path"].toString());
        m_listWidget->addItem(item);
    }
}

void LibraryView::onItemClicked(QListWidgetItem *item) {
    if (!item) return;
    QString videoId = item->data(Qt::UserRole).toString();
    emit videoSelected(videoId);
}
