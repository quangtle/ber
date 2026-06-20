#pragma once

#include <QWidget>
#include <QMediaPlayer>
#include <QAudioOutput>
#include <QVideoWidget>
#include <QSlider>
#include <QLabel>
#include <QPushButton>
#include <QTimer>

class VideoPlayer : public QWidget {
    Q_OBJECT

public:
    explicit VideoPlayer(QWidget *parent = nullptr);
    void load(const QString &url);

signals:
    void playPauseToggled(bool playing);
    void mouseActivity();
    void fullscreenToggled(bool fullscreen);

public slots:
    void play();
    void pause();
    void togglePlayPause();
    void seek(int seconds);
    void setVolume(int volume);
    void toggleMute();

protected:
    bool eventFilter(QObject *obj, QEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;
    void enterEvent(QEnterEvent *event) override;

private slots:
    void onPositionChanged(qint64 position);
    void onDurationChanged(qint64 duration);
    void onStateChanged(QMediaPlayer::PlaybackState state);

private:
    void setupUi();
    void resetHideTimer();
    static QString formatTime(qint64 ms);

    QMediaPlayer *m_mediaPlayer;
    QAudioOutput *m_audioOutput;
    QVideoWidget *m_videoWidget;

    QPushButton *m_playPauseBtn;
    QSlider *m_seekBar;
    QLabel *m_timeLabel;
    QLabel *m_durationLabel;
    QPushButton *m_muteBtn;
    QSlider *m_volumeSlider;
    QWidget *m_controlsBar;
    QTimer *m_hideTimer;

    bool m_seekDragging = false;
};
