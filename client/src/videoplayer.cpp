#include "videoplayer.h"

#include <QVBoxLayout>
#include <QMouseEvent>

VideoPlayer::VideoPlayer(QWidget *parent)
    : QWidget(parent)
    , m_mediaPlayer(new QMediaPlayer(this))
    , m_audioOutput(new QAudioOutput(this))
    , m_videoWidget(new QVideoWidget(this))
{
    auto *layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);
    layout->addWidget(m_videoWidget);

    m_mediaPlayer->setVideoOutput(m_videoWidget);
    m_mediaPlayer->setAudioOutput(m_audioOutput);
    m_audioOutput->setVolume(1.0);

    m_videoWidget->installEventFilter(this);

    connect(m_mediaPlayer, &QMediaPlayer::positionChanged, this, &VideoPlayer::positionChanged);
    connect(m_mediaPlayer, &QMediaPlayer::durationChanged, this, &VideoPlayer::durationChanged);
    connect(m_mediaPlayer, &QMediaPlayer::playbackStateChanged, this, [this](QMediaPlayer::PlaybackState state) {
        emit playPauseToggled(state == QMediaPlayer::PlayingState);
    });
    connect(m_mediaPlayer, &QMediaPlayer::mediaStatusChanged, this, [this](QMediaPlayer::MediaStatus status) {
        if (status == QMediaPlayer::LoadedMedia || status == QMediaPlayer::BufferedMedia)
            emit mediaReady();
    });
}

void VideoPlayer::load(const QString &url) {
    m_mediaPlayer->setSource(QUrl(url));
    m_mediaPlayer->play();
}

qint64 VideoPlayer::position() const { return m_mediaPlayer->position(); }

void VideoPlayer::play() { m_mediaPlayer->play(); }
void VideoPlayer::pause() { m_mediaPlayer->pause(); }
void VideoPlayer::stop() {
    m_mediaPlayer->pause();
    m_mediaPlayer->setSource(QUrl());
}

void VideoPlayer::togglePlayPause() {
    if (m_mediaPlayer->playbackState() == QMediaPlayer::PlayingState)
        m_mediaPlayer->pause();
    else
        m_mediaPlayer->play();
}

void VideoPlayer::seek(int seconds) { m_mediaPlayer->setPosition(seconds * 1000); }

void VideoPlayer::setVolume(int volume) {
    m_audioOutput->setVolume(volume / 100.0);
}

void VideoPlayer::toggleMute() {
    bool newState = !m_audioOutput->isMuted();
    m_audioOutput->setMuted(newState);
    emit mutedChanged(newState);
}

bool VideoPlayer::eventFilter(QObject *obj, QEvent *event) {
    if (obj == m_videoWidget && event->type() == QEvent::MouseButtonDblClick) {
        auto *w = window();
        bool fs = !w->isFullScreen();
        if (fs) w->showFullScreen(); else w->showNormal();
        emit fullscreenToggled(fs);
        return true;
    }
    return QWidget::eventFilter(obj, event);
}
