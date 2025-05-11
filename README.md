
![HomeNCo-logo](https://github.com/user-attachments/assets/2fd0196a-1147-4c46-a31b-b0088f8e85bf)

---

## Presentation

This project is made by Computer Science students for an assignment involving a client application, in this case a website, and a real time interaction with the physical world via ESP32 boards.

**Home&Co** _(Co for Connect, Control, Company)_ is a project using **ESP32** dev boards and a **Raspberry Pi 3B+** to monitor, control and automate home appliances like the lights, heaters, front door, water valves for watering plants, etc.

Its objectives are to provide **security**, **control**, **comfort**, **peace of mind** and **energy saving** to your home.

### 1. Security

Our front door device will control the access policy of the front door of your home:
- **RFID** sensor needing a **badge** to open the door
- **bell button** ringing the bell and **notifying you on your phone**
- **locking mechanism** controlled **remotely** (no need to give keys to anybody)
- **presence detector** to **monitor** the activity in front of the door
- **house locking function** to **automatically turn off** all the lights and use all presence detectors as **intrusion detectors**

### 2. Control

You can use the website to **remotely control at any time** the _lights_, the _door_, the _heaters_ and the _watering of plants_.

You have the **full control of your house** from any device.

### 3. Comfort

You can **automate the heaters** to activate at any time of the day, to find a warm house when coming back from a hike, vacation or any activity.

### 4. Peace of mind

You won't have to go through your checklist three times or ask your neighbour to water your plants before going out!

You can **monitor your house from any device** using the website in **real time**: from the _presence detectors_ to the _lights_, _heaters_, _humidity of the plants_ and _temperature of any room_.

### 5. Energy saving

No need to keep heating your home when you're not there!

You'll be able to control and monitor the **heaters** remotely and according to your needs and presence at home.

You'll also be able to **monitor the energy of specific high consumption appliances** with our **smart power outlet** and control them remotely, and even give them schedules to be working or not.


## How it works

### Needs analysis

- **Goal**: fully functional home with remote control (from the web browser)
- **Security Constraint**: no direct access to ESP32 microcontrollers with their sensors and actuators, to database, MQTT or backend application.
- **Components**:
  - Raspberry Pi 3 or more for Wi-Fi access point, MQTT, backend app and webserver.
  - ESP32 for sensors & actuators

### Technical choices

- **Golang** for the backend app and webserver with GORM:
	- Golang is an excellent choice for building efficient, concurrent applications due to its simplicity and strong type system.
    - The Go standard library is comprehensive, making it easy to handle tasks like networking, file I/O, and concurrency.
    - GORM, a popular ORM for Golang, simplifies database interactions by providing a clean and intuitive API.

- **PostgreSQL** for the database:
	- PostgreSQL is an advanced, open source relational database management system known for its robustness, scalability, and ACID compliance.
    - It offers features like transactions, indexing, and support for complex data types, making it ideal for handling diverse and large-scale applications.

- **Mosquitto** for the MQTT broker:
	- Mosquitto is a lightweight, open source MQTT broker that's highly reliable and easy to configure.
    - Its simplicity and performance make it an excellent choice for real-time communication between devices and services, especially in scenarios where low latency and high availability are crucial.

- **RaspAP** for the Wi-Fi Access Point in the Raspberry Pi:
  - RaspAP provides a user-friendly interface for setting up and managing a Wi-Fi access point on a Raspberry Pi.
  - It simplifies the process of configuring network settings, security protocols, and other related tasks, making it accessible even to users with limited technical expertise.

- **Websocket** between JavaScript clients and Golang server:
  - Websockets provide full-duplex communication channels over a single TCP connection, enabling real-time data exchange between the frontend and backend.
  - This is particularly useful for applications requiring immediate updates and interactions, such as chat applications or live dashboards.

- **PlatformIO** and **C**/**C++** for the ESP32:
	- PlatformIO offers a unified development environment for embedded systems, making it easy to manage multiple projects and boards.
    - It supports a wide range of platforms and tools, including C/C++, which is ideal for writing efficient and low-level code for microcontrollers like the ESP32.

### Basic features

- Monitoring sensors with the web interface
- Controlling actuators from the web interface
- Data collection in the database
- Light and temperature management

### Additional features (future release)

- Front door monitoring with `RFID` badge
- Energy consumption management
- Shutter and window management
- Fire and smoke detection
- Alarm mode
- Ventilation management
- Scheduler for actuators (heaters, light, shutters, ventilation)
- Statistics and prevision dashboards

### Components architecture

#### Overall Infrastructure

```mermaid
stateDiagram-v2
    S: Raspberry Pi Server
    state S {
        MQTT: MQTT Broker
        A: Server Application
        DB: PostgreSQL Database
        
        A --> MQTT: Monitors & send commands
        MQTT --> A: Notify
        
        A --> DB: Updates the database
    }
    
    LD: Light Microcontroller (ESP32)
    state LD {
        App: Application
        B: Broker
        LC: Light Controller
        LS: Luminosity Sensor
        TS: Temperature Sensor
        PD: Presence Detector
        
        App --> B: Instanciates
        App --> LC: Instanciates
        App --> LS: Instanciates
        App --> TS: Instanciates
        App --> PD: Instanciates
        LC --> B: Subscribes channel
        LS --> B: Publishes value
        TS --> B: Publishes value
        PD --> B: Publishes value
    }
    
    FW: Firewall
    I: Internet
    
    B --> MQTT: Connects
    FW --> A: Protects
    A --> FW: Listens to requests
    FW --> I
```

#### Database

```mermaid
---
title: Entity Relationship Diagram
---
erDiagram
    DATA {
        int id
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
        string device_id
        int user_id
        int module_id
        string module_name
        string module_value
    }
    DEVICES {
        string id
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
        int location_id
        string type
        string name
    }
    LOCATIONS {
        int id
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
        string type
        string name
    }
    MODULES {
        int id
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
        string device_id
        string name
        string value
    }
    USERS {
        int id
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
        string name
        string password_hash
        string email
        string phone_number
        string rfid
        string status
    }
    SESSIONS {
        string token
        string data
        timestamp expiry
    }
    
    LOCATIONS ||--o{ DEVICES : contains
    DEVICES }|--|| MODULES : "is composed of"
    DATA ||--o{ DEVICES : "is about"
    DATA |o--o{ USERS : "owns"
```

#### MQTT

```mermaid
---
title: MQTT topics/channels
---

flowchart LR
    system["`_SYSTEM NAME_`"]
    location_type["`_LOCATION TYPE_`"]
    location_id["`_LOCATION ID_`"]
    device_type["`_DEVICE TYPE_`"]
    device_id["`_DEVICE ID_`"]
    module["`_MODULE_`"]
    
    home["`**home**`"]
    
    kitchen["`**kitchen**`"]
    garden["`**garden**`"]
    room["`**room**`"]
    
    kitchen_id["`**1**`"]
    room1_id["`**2**`"]
    room2_id["`**3**`"]
    garden_id["`**4**`"]
    
    light["`**light**`"]
    water_plant["`**waterPlant**`"]
    
    device1["`**ESP32-af6f-43b1-a20e**`"]
    device2["`**ESP32-db87-47b7-998d**`"]
    device3["`**ESP32-808e-4c62-9c4f**`"]
    device4["`**ESP32-0158-4d61-85c2**`"]
    
    light_controller["`**lightController**`"]
    luminosity_sensor["`**luminositySensor**`"]
    water_valve["`**waterValve**`"]
    humidity_sensor["`**humiditySensor**`"]
    
    subgraph OUR SYSTEM NAME
        system
        home
    end
    subgraph LOCATION
        subgraph  
        location_type
        kitchen
        garden
        room
        end
        subgraph  
        location_id
        kitchen_id
        garden_id
        room1_id
        room2_id
        end
    end
    subgraph DEVICE 
       subgraph  
        device_type
        light
        water_plant
        end
        subgraph  
        device_id
        device1
        device2
        device3
        device4
        end 
    end
    subgraph SPECIFIC CHANNEL
        module
        light_controller
        luminosity_sensor
        water_valve
        humidity_sensor
    end

    system  --- location_type   --- location_id     --- device_type     --- device_id   --- module
    home    --- kitchen     --- kitchen_id  --- light           --- device1     --- light_controller & luminosity_sensor
    home    --- garden      --- garden_id   --- water_plant     --- device2     --- water_valve & humidity_sensor
    home    --- room        --- room1_id    --- light           --- device3     --- light_controller & luminosity_sensor
    home    --- room        --- room2_id    --- light           --- device4     --- light_controller & luminosity_sensor
```

#### Web Server file architecture

```
├── cmd
│   └── web
│       ├── context.go
│       ├── handlers.go
│       ├── helpers.go
│       ├── main.go
│       ├── middleware.go
│       ├── models.go
│       ├── routes.go
│       ├── server.go
│       ├── templates.go
│       └── websocket.go
├── go.mod
├── go.sum
├── internal
│   ├── data
│   │   ├── broker.go
│   │   ├── consumption-sensor.go
│   │   ├── data.go
│   │   ├── device.go
│   │   ├── imodule.go
│   │   ├── light-controller.go
│   │   ├── light-sensor.go
│   │   ├── location.go
│   │   ├── luminosity-sensor.go
│   │   ├── models.go
│   │   ├── module.go
│   │   ├── presence-detector.go
│   │   ├── reset.go
│   │   ├── startup.go
│   │   ├── subscription.go
│   │   ├── temperature-sensor.go
│   │   └── value-conversion.go
│   ├── mailer
│   │   ├── mailer.go
│   │   └── templates
│   │       └── alert-notification.tmpl
│   └── validator
│       └── validator.go
├── README.md
├── ui
│   ├── assets
│   │   ├── css
│   │   │   ├── base.scss
│   │   │   └── style.scss
│   │   ├── font
│   │   │   ├── Dosis-VariableFont_wght.ttf
│   │   └── img
│   │       └── logo
│   │           └── logo.png
│   ├── efs.go
│   └── templates
│       ├── base.tmpl
│       ├── pages
│       │   ├── dashboard.tmpl
│       │   ├── error.tmpl
│       │   └── home.tmpl
│       └── partials
│           └── partial-example.tmpl
└── vendor
```

#### ESP32 C++ File Architecture

```
├── include
│   ├── Application.h
│   ├── Broker.h
│   ├── ConsumptionSensor.h
│   ├── environment.h
│   ├── IModule.h
│   ├── IObservable.h
│   ├── IObserver.h
│   ├── LightController.h
│   ├── LightSensor.h
│   ├── LuminositySensor.h
│   ├── ModuleFactory.h
│   ├── MyAny.h
│   ├── PresenceDetector.h
│   ├── README
│   ├── TemperatureSensor.h
│   └── utils.h
├── lib
│   ├── README
│   └── xht11
│       ├── xht11.cpp
│       └── xht11.h
├── LICENSE
├── platformio.ini
├── README.md
└── src
   ├── Application.cpp
   ├── Broker.cpp
   ├── ConsumptionSensor.cpp
   ├── HomeIoT.ino
   ├── LightController.cpp
   ├── LightSensor.cpp
   ├── LuminositySensor.cpp
   ├── ModuleFactory.cpp
   ├── PresenceDetector.cpp
   ├── TemperatureSensor.cpp
   └── utils.cpp
```

#### Class diagram

```mermaid
---
title: Device class diagram
---
classDiagram
    class Application {
        # Application *app$
        # String location
        # unsigned int locationID
        # String root_topic
        # bool wait_for_setup
        # unsigned int publish_interval
        # unsigned long lastPublishTime
        # WiFiClient network
        # Broker *broker
        # IModule *lightController
        # IModule *lightSensor
        # IModule *luminositySensor
        # IModule *presenceDetector
        # IModule *temperatureSensor
        # IModule *consumptionSensor
        
        # isWaitingForSetup() bool
        # onSetupMessage(char payload[]) void
        # reset() void
        # setupModule(const char* name, const char* value) void
        # messageHandler(MQTTClient *client, char topic[], char payload[], int length) void$
        # unsubscribeAllTopics() void
        # setRootTopic() void

        + getInstance() Application*$
        + Application() Application
        + brokerLoop() void
        + startup() void
        + init(WiFiClient wifi) void
        + sensorLoop() void$
    }
    class Broker {
        # MQTTClient mqtt
        # WiFiClient wifi
        # String root_topic

        # Broker(WiFiClient network) Broker

        + newBroker(WiFiClient network, void cb(MQTTClient *client, char topic[], char bytes[], int length)) Broker*$
        + sub(const String &module_name) void
        + pub(const String &module_name, const String &value) void
        + unsub(const String &module_name) void
        + setRootTopic(const String &topic) void
        + loop() void
    }
    class IModule {
        <<interface>>
        # Broker *broker
        # String name
        + setValue(const char * value) void*
        + getValue() const String*
        + getValueReference() const void**
        + getName() String
    }
    class ModuleFactory {
        + newModule(Broker *broker, String type) IModule*$
    }
    class IObservable {
        <<interface>>
        + ~IObservable()*
        + Attach(IObserver *observer) void*
        + Detach(IObserver *observer) void*
        + Notify() void*
    }
    class IObserver {
        <<interface>>
        + ~IObserver()*
        + Update(const String &value) void*
    }
    class LightController {
        # Broker *broker
        # String name
        # bool value
        
        + LightController(Broker *broker, bool value) LightController
        + setValue(const char * value) void
        + getValue() const String
        + getValueReference() const * void
        + getName() String
        + Attach(IObserver *observer) void
        + Detach(IObserver *observer) void
        + Notify() void
        + Update(const String &value) void
    }
    class PresenceDetector {
        # Broker *broker
        # String name
        # bool value

        + PresenceDetector(Broker *broker, bool value) PresenceDetector
        + setValue(const char * value) void
        + getValue() const String
        + getValueReference() const * void
        + getName() String
        + Attach(IObserver *observer) void
        + Detach(IObserver *observer) void
        + Notify() void
        + Update(const String &value) void
    }
    Application *.. IModule
    ModuleFactory ..> IModule
    Application --> ModuleFactory
    Application *.. Broker
    LightController ..|> IModule
    PresenceDetector ..|> IModule
    IObservable *.. IObserver
    IModule ..|> IObserver
    IModule ..|> IObservable
```

#### Startup / Setup

##### Startup/Setup message
```json
{
  "id": "ESP32-af6f-43b1-a20e",
  "type": "light",
  "location_id": 3,
  "location_type": "room",
  "location_name": "room 3",
  "modules": [
    {
      "name": "lightController",
      "value": "False"
    },
    {
      "name": "lightSensor",
      "value": "True"
    },
    {
      "name": "luminositySensor",
      "value": "150.0"
    },
    {
      "name": "presenceDetector",
      "value": "True"
    },
    {
      "name": "temperatureSensor",
      "value": "22.5"
    },
    {
      "name": "consumptionSensor",
      "value": "32.45"
    }
  ]
}
```

##### Setup sequence diagram

```mermaid
sequenceDiagram
    autonumber
    box Teal Device
    participant S as Setup
    participant A as Application
    participant B as Broker
    participant LiC as LightController
    participant LiS as LightSensor
    participant PD as PresenceDetector
    participant LuS as LuminositySensor
    participant TS as TemperatureSensor
    participant CS as ConsumptionSensor
    end
    box Green Server
    participant MQTT as MQTT Broker
    participant Server as Go Application
    participant DB
    end
    S ->> A: Instanciate
    A ->> B: Instanciate
    B ->> MQTT: Connect
    A ->> B: Prepare `startup` message
    B ->> MQTT: Send `startup` message
    MQTT ->> Server: Relay `startup` message from Device
    Server ->> DB: Check if Device exists and creates it if necessary
    DB ->> Server: Send Device data back to prepare `setup` message
    Server ->> MQTT: Send `setup` message
    MQTT ->> B: Relay `setup` message
    B ->> A: Parse `setup` message
    A ->> A: Update `location` and `locationID`
    A ->> LiC: Update value
    A ->> LiS: Update value
    A ->> PD: Update value
    A ->> LuS: Update value
    A ->> TS: Update value
    A ->> CS: Update value
    A ->> S: Setup complete
```

