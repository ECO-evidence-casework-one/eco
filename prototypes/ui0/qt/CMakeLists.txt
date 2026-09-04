cmake_minimum_required(VERSION 3.21)
project(eco_ui0_qt LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Qt6 6.6 REQUIRED COMPONENTS Quick QuickControls2)
qt_standard_project_setup(REQUIRES 6.6)

qt_add_executable(ECO_UI0_QT
    main.cpp
)

qt_add_qml_module(ECO_UI0_QT
    URI Eco.Ui0
    VERSION 1.0
    QML_FILES Main.qml
)

target_link_libraries(ECO_UI0_QT PRIVATE Qt6::Quick Qt6::QuickControls2)

set_target_properties(ECO_UI0_QT PROPERTIES
    WIN32_EXECUTABLE TRUE
)

install(TARGETS ECO_UI0_QT RUNTIME DESTINATION .)
