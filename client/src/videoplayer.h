#pragma once

#include <QWidget>
#include <QMediaPlayer>
#include <QVideoWidget>

class VideoPlayer : public QWidget {
    Q_OBJECT

public:
    explicit VideoPlayer(QWidget *parent = nullptr);
    void load(const QString &url);

public slots:
    void play();
    void pause();
    void togglePlayPause();

private:
    void setupUi();

    QMediaPlayer *m_mediaPlayer;
    QVideoWidget *m_videoWidget;
};
