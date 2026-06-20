#pragma once

#include <QWidget>
#include <QMediaPlayer>
#include <QAudioOutput>
#include <QVideoWidget>
#include <QSlider>
#include <QLabel>
#include <QPushButton>
#include <QTimer>
#include <QEvent>

class VideoPlayer : public QWidget {
    Q_OBJECT

public:
    explicit VideoPlayer(QWidget *parent = nullptr);
    void load(const QString &url);

protected:
    bool eventFilter(QObject *obj, QEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;
    void enterEvent(QEnterEvent *event) override;
    void resizeEvent(QResizeEvent *event) override;
    void showEvent(QShowEvent *event) override;

public slots:
    void play();
    void pause();
    void togglePlayPause();
    void seek(int seconds);
    void setVolume(int volume);
    void toggleMute();

signals:
    void playPauseToggled(bool playing);
    void mouseActivity();
    void fullscreenToggled(bool fullscreen);

private slots:
    void onPositionChanged(qint64 position);
    void onDurationChanged(qint64 duration);
    void onStateChanged(QMediaPlayer::PlaybackState state);

private:
    void setupUi();
    void showControls();
    void hideControls();
    void resetHideTimer();
    static QString formatTime(qint64 ms);

    QMediaPlayer *m_mediaPlayer;
    QAudioOutput *m_audioOutput;
    QVideoWidget *m_videoWidget;

    // controls
    QPushButton *m_playPauseBtn;
    QSlider *m_seekBar;
    QLabel *m_timeLabel;
    QLabel *m_durationLabel;
    QPushButton *m_muteBtn;
    QSlider *m_volumeSlider;
    QWidget *m_controlsOverlay;
    QTimer *m_hideTimer;

    bool m_seekDragging = false;
};
