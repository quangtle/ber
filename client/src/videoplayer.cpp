#include "videoplayer.h"

#include <QVBoxLayout>

VideoPlayer::VideoPlayer(QWidget *parent)
    : QWidget(parent)
    , m_mediaPlayer(new QMediaPlayer(this))
    , m_videoWidget(new QVideoWidget(this))
{
    setupUi();
    m_mediaPlayer->setVideoOutput(m_videoWidget);
}

void VideoPlayer::setupUi() {
    auto *layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);
    layout->addWidget(m_videoWidget);
}

void VideoPlayer::load(const QString &url) {
    m_mediaPlayer->setSource(QUrl(url));
    play();
}

void VideoPlayer::play() {
    m_mediaPlayer->play();
}

void VideoPlayer::pause() {
    m_mediaPlayer->pause();
}

void VideoPlayer::togglePlayPause() {
    if (m_mediaPlayer->playbackState() == QMediaPlayer::PlayingState) {
        m_mediaPlayer->pause();
    } else {
        m_mediaPlayer->play();
    }
}
