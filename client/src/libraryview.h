#pragma once

#include <QWidget>
#include <QListWidget>
#include <QList>
#include <QJsonObject>
#include <QNetworkAccessManager>

class ApiClient;

class LibraryView : public QWidget {
    Q_OBJECT

public:
    explicit LibraryView(ApiClient *client, QWidget *parent = nullptr);

    void refresh();

signals:
    void videoSelected(const QString &videoId);

private slots:
    void onItemClicked(QListWidgetItem *item);
    void onLibraryLoaded(const QList<QJsonObject> &videos);

private:
    void setupUi();
    void loadThumbnail(const QString &videoId, QListWidgetItem *item);

    ApiClient *m_apiClient;
    QListWidget *m_gridWidget;
    QNetworkAccessManager *m_nam;
};
