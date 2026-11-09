package co.edu.unicauca.reacciones;

import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;

@SpringBootTest(properties = {"spring.cloud.consul.enabled=false"})
class ServidorReaccionesApplicationTests {

	@Test
	void contextLoads() {
	}
}
