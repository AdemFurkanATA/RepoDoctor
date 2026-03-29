package app.service;

import java.util.ArrayList;
import java.util.List;

public class UserService {
    public List<String> listUsers() {
        List<String> users = new ArrayList<>();
        users.add("Ada");
        users.add("Linus");
        return users;
    }
}
