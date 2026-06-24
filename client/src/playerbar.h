#pragma once

#include <QWidget>
#include <QPushButton>
#include <QLabel>
#include <QSlider>

class VideoPlayer;
class ApiClient;

class PlayerBar : public QWidget {
    Q_OBJECT

public:
    explicit PlayerBar(ApiClient *client, QWidget *parent = nullptr);
    void playVideo(const QString &videoId, qint64 startPosMs = 0);
    void stop();

signals:
    void closeClicked();

private slots:
    void onPlayPause();
    void onVolumeChanged(int volume);
    void onMuteToggle();
    void onPlayerPositionChanged(qint64 pos);
    void onPlayerDurationChanged(qint64 dur);

private:
    void setupUi();
    static QString formatTime(qint64 ms);

    ApiClient *m_apiClient;
    VideoPlayer *m_videoPlayer;
    QPushButton *m_playPauseBtn;
    QSlider *m_seekBar;
    QLabel *m_timeLabel;
    QLabel *m_durationLabel;
    QPushButton *m_muteBtn;
    QSlider *m_volumeSlider;
    QLabel *m_titleLabel;
    QPushButton *m_closeBtn;
    QString m_currentVideoId;
    qint64 m_pendingSeekMs = 0;
    bool m_seekDragging = false;
};
