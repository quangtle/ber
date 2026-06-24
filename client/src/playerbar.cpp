#include "playerbar.h"
#include "videoplayer.h"
#include "apiclient.h"

#include <QHBoxLayout>
#include <QVBoxLayout>
#include <QJsonObject>

PlayerBar::PlayerBar(ApiClient *client, QWidget *parent)
    : QWidget(parent)
    , m_apiClient(client)
    , m_videoPlayer(new VideoPlayer(this))
    , m_playPauseBtn(new QPushButton("▶"))
    , m_seekBar(new QSlider(Qt::Horizontal))
    , m_timeLabel(new QLabel("0:00"))
    , m_durationLabel(new QLabel("0:00"))
    , m_muteBtn(new QPushButton("🔊"))
    , m_volumeSlider(new QSlider(Qt::Horizontal))
    , m_titleLabel(new QLabel())
    , m_closeBtn(new QPushButton("✕"))
{
    setupUi();

    connect(m_closeBtn, &QPushButton::clicked, this, &PlayerBar::closeClicked);
    connect(m_playPauseBtn, &QPushButton::clicked, this, &PlayerBar::onPlayPause);
    connect(m_seekBar, &QSlider::sliderPressed, this, [this]() { m_seekDragging = true; });
    connect(m_seekBar, &QSlider::sliderReleased, this, [this]() {
        m_seekDragging = false;
        m_videoPlayer->seek(m_seekBar->value());
    });
    connect(m_volumeSlider, &QSlider::valueChanged, this, &PlayerBar::onVolumeChanged);
    connect(m_muteBtn, &QPushButton::clicked, this, &PlayerBar::onMuteToggle);

    // Track the video player state
    connect(m_videoPlayer, &VideoPlayer::positionChanged, this, &PlayerBar::onPlayerPositionChanged);
    connect(m_videoPlayer, &VideoPlayer::durationChanged, this, &PlayerBar::onPlayerDurationChanged);
    connect(m_videoPlayer, &VideoPlayer::mediaReady, this, [this]() {
        if (m_pendingSeekMs > 0) {
            m_videoPlayer->seek(static_cast<int>(m_pendingSeekMs / 1000));
            m_pendingSeekMs = 0;
        }
    });
    connect(m_videoPlayer, &VideoPlayer::playPauseToggled, this, [this](bool playing) {
        m_playPauseBtn->setText(playing ? "⏸" : "▶");
    });
    connect(m_videoPlayer, &VideoPlayer::mutedChanged, this, [this](bool muted) {
        m_muteBtn->setText(muted ? "🔇" : (m_volumeSlider->value() == 0 ? "🔇" : "🔊"));
    });
}

void PlayerBar::setupUi() {
    auto *layout = new QHBoxLayout(this);
    layout->setContentsMargins(8, 4, 8, 4);

    m_videoPlayer->setMaximumSize(356, 200);
    m_videoPlayer->setMinimumSize(160, 90);
    layout->addWidget(m_videoPlayer);

    auto *rightCol = new QVBoxLayout();
    rightCol->setSpacing(4);

    auto *topRow = new QHBoxLayout();
    m_titleLabel->setStyleSheet("font-size: 14px; font-weight: bold;");
    topRow->addWidget(m_titleLabel, 1);
    m_closeBtn->setFixedWidth(28);
    topRow->addWidget(m_closeBtn);
    rightCol->addLayout(topRow);

    auto *ctrlRow = new QHBoxLayout();
    ctrlRow->setSpacing(6);

    m_playPauseBtn->setFixedWidth(36);
    ctrlRow->addWidget(m_playPauseBtn);

    m_timeLabel->setStyleSheet("font-size: 12px;");
    ctrlRow->addWidget(m_timeLabel);

    m_seekBar->setMinimumWidth(100);
    ctrlRow->addWidget(m_seekBar, 1);

    m_durationLabel->setStyleSheet("font-size: 12px;");
    ctrlRow->addWidget(m_durationLabel);

    m_muteBtn->setFixedWidth(36);
    ctrlRow->addWidget(m_muteBtn);

    m_volumeSlider->setRange(0, 100);
    m_volumeSlider->setValue(100);
    m_volumeSlider->setFixedWidth(100);
    ctrlRow->addWidget(m_volumeSlider);

    rightCol->addLayout(ctrlRow);
    layout->addLayout(rightCol, 1);
}

void PlayerBar::playVideo(const QString &videoId, qint64 startPosMs) {
    m_currentVideoId = videoId;
    m_pendingSeekMs = startPosMs;
    m_seekBar->setValue(0);
    m_timeLabel->setText("0:00");
    m_durationLabel->setText("0:00");
    m_playPauseBtn->setText("⏸");
    m_videoPlayer->load(m_apiClient->streamUrl(videoId));
    m_titleLabel->setText("Loading...");

    m_apiClient->fetchVideoInfo(videoId, [this](const QJsonObject &info) {
        m_titleLabel->setText(info["title"].toString());
    });
}

void PlayerBar::stop() {
    m_videoPlayer->stop();
    m_seekBar->setValue(0);
    m_timeLabel->setText("0:00");
    m_durationLabel->setText("0:00");
    m_titleLabel->clear();
}

void PlayerBar::onPlayPause() {
    m_videoPlayer->togglePlayPause();
}

void PlayerBar::onVolumeChanged(int volume) {
    m_videoPlayer->setVolume(volume);
    m_muteBtn->setText(volume == 0 ? "🔇" : "🔊");
}

void PlayerBar::onMuteToggle() {
    m_videoPlayer->toggleMute();
}

void PlayerBar::onPlayerPositionChanged(qint64 pos) {
    if (!m_seekDragging)
        m_seekBar->setValue(static_cast<int>(pos / 1000));
    m_timeLabel->setText(formatTime(pos));
}

void PlayerBar::onPlayerDurationChanged(qint64 dur) {
    m_seekBar->setRange(0, static_cast<int>(dur / 1000));
    m_durationLabel->setText(formatTime(dur));
}

QString PlayerBar::formatTime(qint64 ms) {
    int secs = static_cast<int>(ms / 1000);
    int min = secs / 60;
    return QString("%1:%2").arg(min).arg(secs % 60, 2, 10, QChar('0'));
}
