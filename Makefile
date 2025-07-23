COVER_PKGS = "./handlers,./middleware,./models"

COVER_PROFILE = coverage.out

default: help

test-coverage:
	@echo "--> Gerando relatório de cobertura de testes..."
	# Executa os testes, medindo a cobertura apenas dos pacotes especificados e salvando o resultado.
	go test -coverpkg=$(COVER_PKGS) -coverprofile=$(COVER_PROFILE) ./...
	@echo "--> Abrindo relatório no navegador..."
	# Abre o relatório HTML interativo.
	go tool cover -html=$(COVER_PROFILE)

help:
	@echo ""
	@echo "Comandos disponíveis:"
	@echo " make test-coverage    Gera e exibe o relatório de cobertura de testes."
	@echo ""