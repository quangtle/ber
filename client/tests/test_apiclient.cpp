#include <catch2/catch_test_macros.hpp>
#include <QCoreApplication>
#include "apiclient.h"

TEST_CASE("ApiClient initial state", "[apiclient]") {
    int argc = 0;
    QCoreApplication app(argc, nullptr);
    ApiClient client;

    REQUIRE(client.isConnected() == false);
}

TEST_CASE("ApiClient stream URL generation", "[apiclient]") {
    ApiClient client;
    QString url = client.streamUrl("abc-123");
    REQUIRE(url.contains("abc-123"));
    REQUIRE(url.contains("/api/stream/"));
}

TEST_CASE("ApiClient library URL construction", "[apiclient]") {
    ApiClient client;
    Q_UNUSED(client)
}
