cmake_minimum_required(VERSION 3.16)
project(logriff-c LANGUAGES C)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_EXTENSIONS OFF)

add_executable(logriff src/main.c)

if(MSVC)
    target_compile_options(logriff PRIVATE /W4)
else()
    target_compile_options(logriff PRIVATE -Wall -Wextra -Wpedantic)
endif()
