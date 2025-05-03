FROM mysql:8.0-bookworm

COPY conf/my.cnf /etc/my.cnf.d/my.cnf
COPY conf/mysql-client.cnf /etc/my.cnf.d/mysql-client.cnf

EXPOSE 3306 33060
CMD ["mysqld"]