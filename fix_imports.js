const fs = require('fs');
const path = require('path');

function walk(dir) {
    let results = [];
    const list = fs.readdirSync(dir);
    list.forEach(function(file) {
        file = dir + '/' + file;
        const stat = fs.statSync(file);
        if (stat && stat.isDirectory()) { 
            if (!file.includes('.git') && !file.includes('node_modules') && !file.includes('frontend-next')) {
                results = results.concat(walk(file));
            }
        } else { 
            if (file.endsWith('.go')) results.push(file);
        }
    });
    return results;
}

const files = walk('.');

files.forEach(file => {
    let content = fs.readFileSync(file, 'utf8');
    let originalContent = content;

    // Imports
    content = content.replace(/"chat_distribuido\/controladores\/sockets"/g, '"chat_distribuido/internal/websocket"');
    content = content.replace(/"chat_distribuido\/controladores"/g, '"chat_distribuido/internal/handlers"');
    content = content.replace(/"chat_distribuido\/db"/g, '"chat_distribuido/internal/repository"');
    content = content.replace(/"chat_distribuido\/modelos"/g, '"chat_distribuido/internal/models"');
    content = content.replace(/"chat_distribuido\/middleware"/g, '"chat_distribuido/internal/middleware"');
    content = content.replace(/"chat_distribuido\/utils"/g, '"chat_distribuido/internal/utils"');

    // Packages
    content = content.replace(/^package controladores$/gm, 'package handlers');
    content = content.replace(/^package sockets$/gm, 'package websocket');
    content = content.replace(/^package db$/gm, 'package repository');
    content = content.replace(/^package modelos$/gm, 'package models');

    // Usages
    content = content.replace(/controladores\./g, 'handlers.');
    content = content.replace(/sockets\./g, 'websocket.');
    content = content.replace(/modelos\./g, 'models.');
    
    // DB usages (be careful with local vars named 'db')
    content = content.replace(/db\.ConnectDB/g, 'repository.ConnectDB');
    content = content.replace(/db\.ConnectRedis/g, 'repository.ConnectRedis');
    content = content.replace(/db\.ConnectMinio/g, 'repository.ConnectMinio');
    content = content.replace(/db\.DisconnectDB/g, 'repository.DisconnectDB');
    content = content.replace(/db\.GetCollection/g, 'repository.GetCollection');
    content = content.replace(/db\.RedisClient/g, 'repository.RedisClient');
    content = content.replace(/db\.MinioClient/g, 'repository.MinioClient');
    content = content.replace(/db\.MongoClient/g, 'repository.MongoClient');
    content = content.replace(/db\.GetMinioBucketName/g, 'repository.GetMinioBucketName');

    if (content !== originalContent) {
        fs.writeFileSync(file, content, 'utf8');
        console.log('Updated: ' + file);
    }
});
console.log('Done!');
