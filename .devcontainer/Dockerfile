FROM mcr.microsoft.com/devcontainers/go
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install git sqlite3

RUN go install golang.org/x/tools/gopls@latest

ARG USERNAME=vscode

RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
    && mkdir -p /commandhistory \
    && touch /commandhistory/.bash_history \
    && chown -R $USERNAME /commandhistory \
    && echo "$SNIPPET" >> "/home/$USERNAME/.bashrc"

RUN SNIPPET="autoload -Uz add-zsh-hook; append_history() { fc -W }; add-zsh-hook precmd append_history; export HISTFILE=/commandhistory/.zsh_history" \
    && mkdir -p /commandhistory \
    && touch /commandhistory/.zsh_history \
    && chown -R $USERNAME /commandhistory \
    && echo "$SNIPPET" >> "/home/$USERNAME/.zshrc"

RUN mkdir -p /home/$USERNAME/.claude && chown -R $USERNAME /home/$USERNAME/.claude
RUN chown -R $USERNAME /go

# Set environment variables
ENV DEVCONTAINER=true
ENV SHELL=/bin/zsh
ENV EDITOR=nano
ENV VISUAL=nano
