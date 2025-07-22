# Changelog

## [Unreleased] - 2024-01-XX

### 🆕 New Features
- **🔒 Automatic Security Protection**: Internal endpoints (`/decorators/*`) are now automatically protected
  - `deco.Default()` applies localhost-only protection
  - `deco.DefaultWithSecurity()` for custom security configuration
  - Support for network-based, IP-based, and hostname-based access control

- **🛡️ @Security Decorator**: Network-based access control for application endpoints
  - `@Security(private)` - Allow private networks
  - `@Security(networks="192.168.1.0/24")` - Specific networks
  - `@Security(ips="192.168.1.100")` - Specific IPs

- **🔄 @Proxy Decorator**: Complete API Gateway functionality
  - Service discovery (Consul, DNS, Kubernetes, Static)
  - Load balancing (Round Robin, Least Connections, IP Hash, Weighted)
  - Health checks and circuit breaker pattern

### 🔧 Improvements
- Enhanced documentation with comprehensive API reference
- Improved error handling and logging
- Better integration with existing middleware system

### 🐛 Bug Fixes
- Fixed proxy middleware integration in code generation
- Improved security middleware performance

## Recent Changes

c258ab4 feat(docs): implement automated documentation system with Go Gopher branding (#19)
8fe94eb chore: increment version to 0.6.0
95af94b docs: update documentation for v0.4.0
c25453e chore: increment version to 0.5.0
0635bc7 chore: increment version to 0.4.0
ed8f4c6 docs: update documentation
75de0f0 refactor: rename decorate-gen to deco and improve documentation automation (#18)
fe63d56 chore: increment version to 0.3.0
9f51360 feat: configure binary name as 'deco' and fix documentation generation (#17)
ac922f1 chore: increment version to 0.2.0
