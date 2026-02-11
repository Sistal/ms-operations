CREATE TABLE "Notificacion" (
	id_notificacion serial4 NOT NULL,
	id_usuario int4 NOT NULL,
	titulo varchar(100) NOT NULL,
	cuerpo text NOT NULL,
	leido bool DEFAULT false NOT NULL,
	fecha timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT "Notificacion_pkey" PRIMARY KEY (id_notificacion),
	CONSTRAINT "FK_Notificacion_id_usuario" FOREIGN KEY (id_usuario) REFERENCES "Usuario"(id_usuario)
);
