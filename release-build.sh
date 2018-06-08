#!/bin/bash
rm -rf bin

mkdir bin

xgo --targets=windows/*,linux/386,linux/amd64,linux/arm64 -out bin/lit .
xgo --targets=windows/*,linux/386,linux/amd64,linux/arm64 -out bin/lit-af ./cmd/lit-af

rm -rf out
mkdir out

mkdir out/lit-windows-x64
mkdir out/lit-windows-x86
mkdir out/lit-linux-x64
mkdir out/lit-linux-x86
mkdir out/lit-linux-arm64

mv bin/lit-windows-4.0-amd64.exe out/lit-windows-x64/lit.exe
mv bin/lit-windows-4.0-386.exe out/lit-windows-x86/lit.exe
mv bin/lit-linux-amd64 out/lit-linux-x64/lit
mv bin/lit-linux-386 out/lit-linux-x86/lit
mv bin/lit-linux-arm64 out/lit-linux-arm64/lit
mv bin/lit-af-windows-4.0-amd64.exe out/lit-windows-x64/lit-af.exe
mv bin/lit-af-windows-4.0-386.exe out/lit-windows-x86/lit-af.exe
mv bin/lit-af-linux-amd64 out/lit-linux-x64/lit-af
mv bin/lit-af-linux-386 out/lit-linux-x86/lit-af
mv bin/lit-af-linux-arm64 out/lit-linux-arm64/lit-af

zip -j -9 out/lit-windows-x64.zip out/lit-windows-x64/*
zip -j -9 out/lit-windows-x86.zip out/lit-windows-x86/*
zip -j -9 out/lit-linux-x64.zip out/lit-linux-x64/*
zip -j -9 out/lit-linux-x86.zip out/lit-linux-x86/*
zip -j -9 out/lit-linux-arm64.zip out/lit-linux-arm64/*

rm -rf out/lit-windows-x64
rm -rf out/lit-windows-x86
rm -rf out/lit-linux-x64
rm -rf out/lit-linux-x86
rm -rf out/lit-linux-arm64

rm -rf bin