#pragma once

#include <QWidget>
#include <QPushButton>
#include <QLabel>

class VideoPlayer;
class ApiClient;

class PlayerWidget : public QWidget {
    Q_OBJECT

public:
    explicit PlayerWidget(ApiClient *client, QWidget *parent = nullptr);
    void playVideo(const QString &videoId);

signals:
    void backToLibrary();
    void fullscreenToggled(bool fullscreen);

private:
    void setupUi();

    ApiClient *m_apiClient;
    VideoPlayer *m_videoPlayer;
    QPushButton *m_backBtn;
    QLabel *m_titleLabel;
    QWidget *m_toolbar;
    QString m_currentVideoId;
};
