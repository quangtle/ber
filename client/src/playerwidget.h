#pragma once

#include <QWidget>
#include <QVBoxLayout>
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

private:
    void setupUi();
    void updateControls();

    ApiClient *m_apiClient;
    VideoPlayer *m_videoPlayer;
    QPushButton *m_backBtn;
    QPushButton *m_playPauseBtn;
    QLabel *m_titleLabel;
    QString m_currentVideoId;
};
