@echo off
REM ===========================================================================
REM  lancer.bat — Point d'entree double-cliquable du projet quiz-connaissances.
REM  Il se contente d'appeler « lancer.ps1 » (qui contient toute la logique)
REM  en contournant la politique d'execution PowerShell, sinon Windows refuse
REM  parfois d'executer les scripts .ps1 telecharges.
REM ===========================================================================
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0quiz.ps1"
