# Rinha de Backend - 2025 - Go e Redis

Aplicação feita para o desafio da Rinha de Backend 2025.

# Como funciona?
Basicamente os pagamentos recebidos entram em uma fila e são consumidos por workers.
Sempre buscando o processamento com o processador default
É utilizado o redis como banco e o nginx como load balancer
