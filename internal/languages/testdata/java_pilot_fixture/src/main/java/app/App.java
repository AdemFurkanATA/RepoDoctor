package app;

import app.service.UserService;
import java.util.List;

public class App {
    public static void main(String[] args) {
        UserService svc = new UserService();
        List<String> users = svc.listUsers();
        System.out.println(users.size());
    }
}
