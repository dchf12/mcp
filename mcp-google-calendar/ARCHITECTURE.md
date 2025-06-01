# MCP Google Calendar - Architecture Documentation

## Overview

The MCP Google Calendar project implements a Model Context Protocol (MCP) server that provides Google Calendar integration using Clean Architecture principles. The system is built in Go and follows Domain-Driven Design (DDD) patterns with clear separation of concerns across multiple architectural layers.

## Directory Structure Analysis

```
mcp-google-calendar/
├── cmd/                          # Application entry points
│   └── server/                   # Main server application
├── internal/                     # Internal application code
│   ├── domain/                   # Domain layer (entities, value objects)
│   ├── usecase/                  # Use case layer (business logic)
│   ├── interfaces/               # Interface layer (MCP tools, adapters)
│   └── infrastructure/           # Infrastructure layer (external systems)
│       ├── gcal/                 # Google Calendar API adapter
│       └── repository/           # Data persistence
└── pkg/                          # Public packages
    ├── config/                   # Configuration management
    └── errors/                   # Error definitions
```

## Clean Architecture Implementation

The project follows Clean Architecture with four distinct layers:

### 1. Domain Layer (`internal/domain/`)
- **Entities**: `Calendar`, `Event`
- **Value Objects**: `DateTime`
- **Business Rules**: Validation logic embedded in entities
- **Zero Dependencies**: No external dependencies

### 2. Use Case Layer (`internal/usecase/`)
- **Use Cases**: `CreateEventUseCase`, `GetCalendarsUseCase`
- **Repository Interfaces**: `EventRepository`, `CalendarRepository`
- **Business Logic**: Input validation and orchestration
- **Dependencies**: Only depends on domain layer

### 3. Interface Layer (`internal/interfaces/`)
- **MCP Tools**: `ListCalendarTool`, `CreateEventTool`
- **Tool Registration**: `RegisterCalendarTools`
- **Input/Output Conversion**: JSON marshaling/unmarshaling
- **Dependencies**: Domain and use case layers

### 4. Infrastructure Layer (`internal/infrastructure/`)
- **External APIs**: Google Calendar API adapter
- **Persistence**: Token file repository
- **Cross-cutting Concerns**: Rate limiting, metrics, error handling
- **Dependencies**: Domain layer (implements interfaces)

## Architecture Diagrams

### Layer Architecture Diagram

```mermaid
graph TB
    subgraph "Clean Architecture Layers"
        subgraph "Domain Layer"
            D1[Calendar Entity]
            D2[Event Entity]
            D3[DateTime Value Object]
        end
        
        subgraph "Use Case Layer"
            U1[GetCalendarsUseCase]
            U2[CreateEventUseCase]
            U3[Repository Interfaces]
        end
        
        subgraph "Interface Layer"
            I1[ListCalendarTool]
            I2[CreateEventTool]
            I3[MCP Server Registration]
        end
        
        subgraph "Infrastructure Layer"
            IN1[Google Calendar Adapter]
            IN2[Token File Repository]
            IN3[Rate Limiter]
            IN4[Metrics Collector]
        end
        
        subgraph "External Systems"
            E1[Google Calendar API]
            E2[File System]
            E3[Prometheus Metrics]
        end
    end
    
    %% Dependencies (Inner layers do not depend on outer layers)
    U1 --> D1
    U1 --> D2
    U2 --> D1
    U2 --> D2
    U2 --> D3
    
    I1 --> U1
    I2 --> U2
    
    IN1 --> D1
    IN1 --> D2
    IN1 --> D3
    IN2 --> E2
    
    IN1 --> E1
    IN4 --> E3
    
    %% Interface implementations
    IN1 -.-> U3
    IN2 -.-> U3
```

#### Example Use Case and Interface Interaction

```mermaid
graph TD
    subgraph "Interface Layer"
        I[ListCalendarTool]
    end

    subgraph "Use Case Layer"
        U[GetCalendarsUseCase]
        R[CalendarRepository<br/>interface]
    end

    subgraph "Infrastructure Layer"
        G[GoogleCalendarAdapter]
    end

    subgraph "Domain Layer"
        D[Calendar Entity]
    end

    %% 依存関係
    I -->|uses| U
    U -->|depends on| R
    U -->|uses| D
    G -.->|implements| R
    G -->|creates| D
```

### Component Dependency Diagram

```mermaid
graph LR
    subgraph "Main Application"
        Main[main.go]
    end
    
    subgraph "Configuration"
        Config[config.Config]
        OAuth[oauth.AuthFlow]
        TokenRepo[TokenFileRepo]
    end
    
    subgraph "MCP Server"
        MCPServer[MCP Server]
        ToolReg[Tool Registration]
    end
    
    subgraph "Business Logic"
        GetCalUC[GetCalendarsUseCase]
        CreateEventUC[CreateEventUseCase]
    end
    
    subgraph "Tools"
        ListTool[ListCalendarTool]
        CreateTool[CreateEventTool]
    end
    
    subgraph "Infrastructure"
        GCalAdapter[GoogleCalendarAdapter]
        RateLimit[RateLimiter]
        Metrics[Metrics]
    end
    
    subgraph "External"
        GCalAPI[Google Calendar API]
        PromMetrics[Prometheus]
    end
    
    Main --> Config
    Main --> OAuth
    Main --> MCPServer
    Main --> ToolReg
    
    ToolReg --> ListTool
    ToolReg --> CreateTool
    ToolReg --> GetCalUC
    ToolReg --> CreateEventUC
    ToolReg --> GCalAdapter
    
    ListTool --> GetCalUC
    CreateTool --> CreateEventUC
    
    GetCalUC --> GCalAdapter
    CreateEventUC --> GCalAdapter
    
    GCalAdapter --> RateLimit
    GCalAdapter --> Metrics
    GCalAdapter --> GCalAPI
    
    OAuth --> TokenRepo
    Metrics --> PromMetrics
```

### Data Flow Diagram

```mermaid
sequenceDiagram
    participant MCP as MCP Client
    participant Server as MCP Server
    participant Tool as Calendar Tool
    participant UC as Use Case
    participant Adapter as GCal Adapter
    participant RateLimit as Rate Limiter
    participant API as Google Calendar API
    participant Metrics as Prometheus
    
    MCP->>Server: tool_call request
    Server->>Tool: execute(request)
    
    Tool->>Tool: validate & parse input
    Tool->>UC: execute(params)
    
    UC->>UC: validate business rules
    UC->>Adapter: repository method
    
    Adapter->>RateLimit: check limits
    RateLimit-->>Adapter: allowed/blocked
    
    alt Rate limit OK
        Adapter->>Metrics: record request
        Adapter->>API: Google API call
        API-->>Adapter: response
        Adapter->>Metrics: record duration
        Adapter-->>UC: domain objects
    else Rate limit exceeded
        Adapter->>Metrics: record rate limit hit
        Adapter-->>UC: rate limit error
    end
    
    UC-->>Tool: result/error
    Tool->>Tool: format response
    Tool-->>Server: MCP response
    Server-->>MCP: tool_call result
    
    Note over Metrics: Continuous monitoring
    Metrics->>Metrics: Export to Prometheus
```

### Sequence Diagram - Create Event Flow

```mermaid
sequenceDiagram
    participant Client as MCP Client
    participant Server as MCP Server
    participant CreateTool as CreateEventTool
    participant CreateUC as CreateEventUseCase
    participant Domain as Event Entity
    participant GCalAdapter as GoogleCalendarAdapter
    participant RateLimit as RateLimiter
    participant GCalService as Google Calendar API
    participant Metrics as Metrics Collector
    
    Client->>Server: create_event tool call
    Note over Client,Server: {"calendar_id": "primary", "title": "Meeting", ...}
    
    Server->>CreateTool: Execute(ctx, request)
    CreateTool->>CreateTool: Parse JSON arguments
    CreateTool->>Domain: Create Event entity
    
    CreateTool->>CreateUC: Execute(ctx, calendarID, event)
    CreateUC->>Domain: Validate() business rules
    
    alt Validation successful
        CreateUC->>GCalAdapter: CreateEvent(ctx, calendarID, event)
        GCalAdapter->>Metrics: recordAPIRequest("create_event")
        GCalAdapter->>RateLimit: Allow()
        
        alt Rate limit OK
            GCalAdapter->>GCalService: Events.Insert(calendarID, event)
            GCalService-->>GCalAdapter: Created event response
            GCalAdapter->>Metrics: recordAPIResponseDuration()
            GCalAdapter->>Domain: Convert to domain Event
            GCalAdapter-->>CreateUC: Created event
        else Rate limit exceeded
            GCalAdapter->>Metrics: recordRateLimitHit()
            GCalAdapter-->>CreateUC: RateLimitExceededError
        end
        
        CreateUC->>Domain: Validate() created event
        CreateUC-->>CreateTool: Result/Error
    else Validation failed
        CreateUC-->>CreateTool: Validation error
    end
    
    CreateTool->>CreateTool: Format MCP response
    CreateTool-->>Server: CallToolResult
    Server-->>Client: Tool execution result
```

## Key Design Decisions and Rationale

### 1. Clean Architecture Pattern
- **Decision**: Implement Clean Architecture with clear layer boundaries
- **Rationale**: 
  - Separation of concerns for maintainability
  - Testability through dependency injection
  - Independence from external frameworks
  - Business logic isolation

### 2. Domain-Driven Design (DDD)
- **Decision**: Use domain entities with embedded validation
- **Rationale**:
  - Business rules are centralized in domain objects
  - Rich domain model prevents invalid states
  - Clear ubiquitous language across the application

### 3. Repository Pattern
- **Decision**: Abstract external dependencies through interfaces
- **Rationale**:
  - Testability through mock implementations
  - Flexibility to change external services
  - Clear contracts between layers

### 4. Metrics and Monitoring
- **Decision**: Implement Prometheus metrics throughout the infrastructure layer
- **Rationale**:
  - Observability for production systems
  - Performance monitoring and alerting
  - API quota management

### 5. Rate Limiting
- **Decision**: Implement client-side rate limiting for Google Calendar API
- **Rationale**:
  - Prevent API quota exhaustion
  - Graceful handling of rate limits
  - Improved reliability

### 6. Error Handling Strategy
- **Decision**: Custom error types with context information
- **Rationale**:
  - Detailed error reporting for debugging
  - Proper error categorization
  - Error wrapping for context preservation

### 7. Configuration Management
- **Decision**: Environment-based configuration with validation
- **Rationale**:
  - Security for sensitive credentials
  - Environment-specific deployments
  - Early validation of configuration

### 8. OAuth Token Management
- **Decision**: File-based token storage with automatic refresh
- **Rationale**:
  - Persistent authentication sessions
  - Security through file permissions
  - Simplified deployment

## Dependencies and Data Flow

### External Dependencies
- **Google Calendar API**: Primary integration target
- **MCP Protocol**: Communication framework
- **OAuth2**: Authentication mechanism
- **Prometheus**: Metrics collection
- **File System**: Token persistence

### Internal Dependencies
- Domain layer: No dependencies (core business logic)
- Use case layer: Depends only on domain
- Interface layer: Depends on use case and domain
- Infrastructure layer: Implements use case interfaces

### Cross-Cutting Concerns
- **Logging**: Structured logging with slog
- **Metrics**: Prometheus metrics collection
- **Error Handling**: Custom error types with context
- **Rate Limiting**: Google API quota management
- **Configuration**: Environment-based setup

## Current Implementation Status

### Completed Features
- ✅ Clean Architecture implementation
- ✅ Domain entities with validation
- ✅ Use case orchestration
- ✅ Google Calendar API integration
- ✅ OAuth2 authentication flow
- ✅ Rate limiting implementation
- ✅ Prometheus metrics
- ✅ MCP tool registration
- ✅ Error handling framework
- ✅ Configuration management
- ✅ Token persistence

### Key Capabilities
1. **List Calendars**: Retrieve user's available calendars
2. **Create Events**: Create new calendar events with validation
3. **Rate Limiting**: Prevent API quota exhaustion
4. **Metrics**: Monitor API usage and performance
5. **Error Handling**: Comprehensive error reporting
6. **Authentication**: OAuth2 flow with token refresh

### Testing Coverage
- Unit tests for domain entities
- Use case testing with mocks
- Infrastructure component testing
- Configuration validation testing

## Future Enhancement Opportunities

1. **Additional MCP Tools**: Update events, delete events, list events
2. **Caching Layer**: Reduce API calls for frequently accessed data
3. **Batch Operations**: Support for bulk event operations
4. **Event Subscriptions**: Real-time calendar change notifications
5. **Multi-calendar Support**: Enhanced calendar management features
6. **Configuration UI**: Web interface for easier setup

This architecture provides a solid foundation for a maintainable, testable, and extensible Google Calendar MCP server while following industry best practices for Go applications.