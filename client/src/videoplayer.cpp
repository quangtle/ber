#include "videoplayer.h"

#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QMouseEvent>

VideoPlayer::VideoPlayer(QWidget *parent)
    : QWidget(parent)
    , m_mediaPlayer(new QMediaPlayer(this))
    , m_audioOutput(new QAudioOutput(this))
    , m_videoWidget(new QVideoWidget(this))
    , m_playPauseBtn(new QPushButton)
    , m_seekBar(new QSlider(Qt::Horizontal))
    , m_timeLabel(new QLabel("0:00"))
    , m_durationLabel(new QLabel("0:00"))
    , m_muteBtn(new QPushButton)
    , m_volumeSlider(new QSlider(Qt::Horizontal))
    , m_controlsBar(new QWidget(this))
    , m_hideTimer(new QTimer(this))
{
    setupUi();

    m_mediaPlayer->setVideoOutput(m_videoWidget);
    m_mediaPlayer->setAudioOutput(m_audioOutput);
    m_audioOutput->setVolume(1.0);

    // track mouse on the whole widget and the video child
    m_videoWidget->installEventFilter(this);
    m_videoWidget->setMouseTracking(true);
    setMouseTracking(true);

    m_hideTimer->setSingleShot(true);
    m_hideTimer->setInterval(1000);
    connect(m_hideTimer, &QTimer::timeout, this, [this]() { m_controlsBar->hide(); });

    connect(m_mediaPlayer, &QMediaPlayer::positionChanged, this, &VideoPlayer::onPositionChanged);
    connect(m_mediaPlayer, &QMediaPlayer::durationChanged, this, &VideoPlayer::onDurationChanged);
    connect(m_mediaPlayer, &QMediaPlayer::playbackStateChanged, this, &VideoPlayer::onStateChanged);

    connect(m_seekBar, &QSlider::sliderPressed, this, [this]() { m_seekDragging = true; });
    connect(m_seekBar, &QSlider::sliderReleased, this, [this]() {
        m_seekDragging = false;
        m_mediaPlayer->setPosition(m_seekBar->value() * 1000);
    });

    connect(m_volumeSlider, &QSlider::valueChanged, this, &VideoPlayer::setVolume);
    connect(m_muteBtn, &QPushButton::clicked, this, &VideoPlayer::toggleMute);
    connect(m_playPauseBtn, &QPushButton::clicked, this, &VideoPlayer::togglePlayPause);
}

void VideoPlayer::setupUi() {
    auto *layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);
    layout->setSpacing(0);
    layout->addWidget(m_videoWidget, 1);

    // controls bar — sits below video, collapses when hidden
    auto *ctrlLayout = new QVBoxLayout(m_controlsBar);
    ctrlLayout->setContentsMargins(0, 0, 0, 0);
    ctrlLayout->setSpacing(0);

    auto *seekRow = new QHBoxLayout();
    seekRow->setContentsMargins(8, 4, 8, 0);
    seekRow->addWidget(m_timeLabel);
    seekRow->addWidget(m_seekBar, 1);
    seekRow->addWidget(m_durationLabel);
    ctrlLayout->addLayout(seekRow);

    auto *btnRow = new QHBoxLayout();
    btnRow->setContentsMargins(8, 4, 8, 8);

    m_playPauseBtn->setText("▶");
    m_playPauseBtn->setFixedWidth(36);

    m_muteBtn->setText("🔊");
    m_muteBtn->setFixedWidth(36);

    m_volumeSlider->setRange(0, 100);
    m_volumeSlider->setValue(100);
    m_volumeSlider->setFixedWidth(120);

    btnRow->addWidget(m_playPauseBtn);
    btnRow->addWidget(m_muteBtn);
    btnRow->addWidget(m_volumeSlider);
    btnRow->addStretch();
    ctrlLayout->addLayout(btnRow);

    layout->addWidget(m_controlsBar);
}

void VideoPlayer::load(const QString &url) {
    m_mediaPlayer->setSource(QUrl(url));
    m_timeLabel->setText("0:00");
    m_durationLabel->setText("0:00");
    m_seekBar->setValue(0);
    play();
}

void VideoPlayer::play() { m_mediaPlayer->play(); }
void VideoPlayer::pause() { m_mediaPlayer->pause(); }

void VideoPlayer::togglePlayPause() {
    if (m_mediaPlayer->playbackState() == QMediaPlayer::PlayingState)
        m_mediaPlayer->pause();
    else
        m_mediaPlayer->play();
}

void VideoPlayer::seek(int seconds) { m_mediaPlayer->setPosition(seconds * 1000); }

void VideoPlayer::setVolume(int volume) {
    m_audioOutput->setVolume(volume / 100.0);
    m_muteBtn->setText(volume == 0 ? "🔇" : "🔊");
}

void VideoPlayer::toggleMute() {
    if (m_audioOutput->isMuted()) {
        m_audioOutput->setMuted(false);
        m_muteBtn->setText(m_volumeSlider->value() == 0 ? "🔇" : "🔊");
    } else {
        m_audioOutput->setMuted(true);
        m_muteBtn->setText("🔇");
    }
}

void VideoPlayer::resetHideTimer() {
    m_controlsBar->show();
    m_hideTimer->start();
    emit mouseActivity();
}

void VideoPlayer::onPositionChanged(qint64 position) {
    if (!m_seekDragging)
        m_seekBar->setValue(static_cast<int>(position / 1000));
    m_timeLabel->setText(formatTime(position));
}

void VideoPlayer::onDurationChanged(qint64 duration) {
    m_seekBar->setRange(0, static_cast<int>(duration / 1000));
    m_durationLabel->setText(formatTime(duration));
}

void VideoPlayer::onStateChanged(QMediaPlayer::PlaybackState state) {
    bool playing = (state == QMediaPlayer::PlayingState);
    m_playPauseBtn->setText(playing ? "⏸" : "▶");
    emit playPauseToggled(playing);
}

bool VideoPlayer::eventFilter(QObject *obj, QEvent *event) {
    if (obj == m_videoWidget && event->type() == QEvent::MouseButtonDblClick) {
        auto *w = window();
        bool fs = !w->isFullScreen();
        if (fs) w->showFullScreen(); else w->showNormal();
        emit fullscreenToggled(fs);
        return true;
    }
    if (event->type() == QEvent::MouseMove)
        resetHideTimer();
    return QWidget::eventFilter(obj, event);
}

void VideoPlayer::mouseMoveEvent(QMouseEvent *event) {
    resetHideTimer();
    QWidget::mouseMoveEvent(event);
}

void VideoPlayer::enterEvent(QEnterEvent *event) {
    resetHideTimer();
    QWidget::enterEvent(event);
}

QString VideoPlayer::formatTime(qint64 ms) {
    int secs = static_cast<int>(ms / 1000);
    int min = secs / 60;
    return QString("%1:%2").arg(min).arg(secs % 60, 2, 10, QChar('0'));
}
