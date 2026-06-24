#pragma once

#include <QWidget>
#include <QMediaPlayer>
#include <QAudioOutput>
#include <QVideoWidget>

class VideoPlayer : public QWidget {
    Q_OBJECT

public:
    explicit VideoPlayer(QWidget *parent = nullptr);
    void load(const QString &url);
    qint64 position() const;

signals:
    void playPauseToggled(bool playing);
    void positionChanged(qint64 positionMs);
    void durationChanged(qint64 durationMs);
    void fullscreenToggled(bool fullscreen);
    void mutedChanged(bool muted);
    void mediaReady();

public slots:
    void play();
    void pause();
    void stop();
    void togglePlayPause();
    void seek(int seconds);
    void setVolume(int volume);
    void toggleMute();

private:
    bool eventFilter(QObject *obj, QEvent *event) override;

    QMediaPlayer *m_mediaPlayer;
    QAudioOutput *m_audioOutput;
    QVideoWidget *m_videoWidget;
};
