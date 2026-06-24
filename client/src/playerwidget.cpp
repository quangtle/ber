#include "playerwidget.h"
#include "videoplayer.h"
#include "apiclient.h"

#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QJsonObject>
#include <QMediaPlayer>

PlayerWidget::PlayerWidget(ApiClient *client, QWidget *parent)
    : QWidget(parent)
    , m_apiClient(client)
    , m_videoPlayer(new VideoPlayer(this))
    , m_backBtn(new QPushButton("← Library"))
    , m_titleLabel(new QLabel())
    , m_toolbar(new QWidget(this))
    , m_playPauseBtn(new QPushButton("⏸"))
    , m_seekBar(new QSlider(Qt::Horizontal))
    , m_timeLabel(new QLabel("0:00"))
    , m_durationLabel(new QLabel("0:00"))
    , m_muteBtn(new QPushButton("🔊"))
    , m_volumeSlider(new QSlider(Qt::Horizontal))
{
    setupUi();

    connect(m_backBtn, &QPushButton::clicked, this, &PlayerWidget::backToLibrary);
    connect(m_videoPlayer, &VideoPlayer::fullscreenToggled, this, &PlayerWidget::fullscreenToggled);
    connect(m_videoPlayer, &VideoPlayer::positionChanged, this, &PlayerWidget::onPlayerPositionChanged);
    connect(m_videoPlayer, &VideoPlayer::durationChanged, this, &PlayerWidget::onPlayerDurationChanged);
    connect(m_videoPlayer, &VideoPlayer::playPauseToggled, this, [this](bool playing) {
        m_playPauseBtn->setText(playing ? "⏸" : "▶");
    });
}

void PlayerWidget::setupUi() {
    auto *layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);
    layout->setSpacing(0);

    // Toolbar: back button + title
    auto *tbLayout = new QHBoxLayout(m_toolbar);
    tbLayout->setContentsMargins(8, 4, 8, 4);
    tbLayout->addWidget(m_backBtn);
    m_titleLabel->setStyleSheet("font-size: 16px; font-weight: bold;");
    tbLayout->addWidget(m_titleLabel, 1);

    layout->addWidget(m_toolbar);

    // Video takes remaining space
    layout->addWidget(m_videoPlayer, 1);

    // Controls bar at bottom
    auto *ctrlLayout = new QHBoxLayout();
    ctrlLayout->setContentsMargins(8, 4, 8, 8);
    ctrlLayout->setSpacing(6);

    m_playPauseBtn->setFixedWidth(40);
    ctrlLayout->addWidget(m_playPauseBtn);

    ctrlLayout->addWidget(m_timeLabel);
    ctrlLayout->addWidget(m_seekBar, 1);
    ctrlLayout->addWidget(m_durationLabel);

    m_muteBtn->setFixedWidth(36);
    ctrlLayout->addWidget(m_muteBtn);

    m_volumeSlider->setRange(0, 100);
    m_volumeSlider->setValue(100);
    m_volumeSlider->setFixedWidth(120);
    ctrlLayout->addWidget(m_volumeSlider);

    layout->addLayout(ctrlLayout);

    // Connections
    connect(m_playPauseBtn, &QPushButton::clicked, m_videoPlayer, &VideoPlayer::togglePlayPause);
    connect(m_seekBar, &QSlider::sliderPressed, this, [this]() { m_seekDragging = true; });
    connect(m_seekBar, &QSlider::sliderReleased, this, [this]() {
        m_seekDragging = false;
        m_videoPlayer->seek(m_seekBar->value());
    });
    connect(m_volumeSlider, &QSlider::valueChanged, this, [this](int v) {
        m_videoPlayer->setVolume(v);
        m_muteBtn->setText(v == 0 ? "🔇" : "🔊");
    });
    connect(m_muteBtn, &QPushButton::clicked, m_videoPlayer, &VideoPlayer::toggleMute);
}

void PlayerWidget::playVideo(const QString &videoId) {
    m_currentVideoId = videoId;
    m_videoPlayer->load(m_apiClient->streamUrl(videoId));
    m_titleLabel->setText("Loading...");
    m_playPauseBtn->setText("⏸");

    m_apiClient->fetchVideoInfo(videoId, [this](const QJsonObject &info) {
        m_titleLabel->setText(info["title"].toString());
    });
}

qint64 PlayerWidget::position() const { return m_videoPlayer->position(); }

void PlayerWidget::stop() {
    m_videoPlayer->stop();
    m_seekBar->setValue(0);
    m_timeLabel->setText("0:00");
    m_durationLabel->setText("0:00");
    m_titleLabel->clear();
}

void PlayerWidget::onPlayerPositionChanged(qint64 pos) {
    if (!m_seekDragging)
        m_seekBar->setValue(static_cast<int>(pos / 1000));
    m_timeLabel->setText(formatTime(pos));
}

void PlayerWidget::onPlayerDurationChanged(qint64 dur) {
    m_seekBar->setRange(0, static_cast<int>(dur / 1000));
    m_durationLabel->setText(formatTime(dur));
}

QString PlayerWidget::formatTime(qint64 ms) {
    int secs = static_cast<int>(ms / 1000);
    int min = secs / 60;
    return QString("%1:%2").arg(min).arg(secs % 60, 2, 10, QChar('0'));
}
