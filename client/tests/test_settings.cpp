#include <catch2/catch_test_macros.hpp>
#include <QCoreApplication>
#include "settings.h"

TEST_CASE("Settings read/write", "[settings]") {
    int argc = 0;
    QCoreApplication app(argc, nullptr);
    Settings settings;

    SECTION("default last server is empty") {
        REQUIRE(settings.lastServer().isEmpty());
    }

    SECTION("write and read last server") {
        settings.setLastServer("192.168.1.100:8080");
        REQUIRE(settings.lastServer() == "192.168.1.100:8080");
    }
}
