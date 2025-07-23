%global debug_package %{nil}

Name:           prometheus-slurm-exporter
Version:        0.21
%define rel     4
Release:        %{rel}%{?dist}
Summary:        Prometheus exporter for SLURM metrics
Group:          Monitoring

License:        GPL 3.0
URL:            https://github.com/omula/prometheus-slurm-exporter

Source0:        %{name}-%{version}.tar.bz2
Source1:        prometheus-slurm-exporter.service
Source2:        LICENSE
Source3:        README.md

Requires(pre): shadow-utils

Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd
BuildRequires:  go systemd-rpm-macros

%description
A Prometheus exporter for metrics extracted from the Slurm resource scheduling system.

%prep
%setup -n %{name}-%{version}

%build
make all


%install
install -m 0755 -vd %{buildroot}%{_bindir}
install -m 0755 bin/%{name} %{buildroot}%{_bindir}

install -m 0755 -vd %{buildroot}%{_unitdir}/
install -m 0755 -vd  %{buildroot}/%{_sharedstatedir}/prometheus
install -m 644 %{SOURCE1} %{buildroot}/%{_unitdir}/%{name}.service

install -m 0755 -vd %{buildroot}%{_datadir}/%{name}
install -m 644 %{SOURCE2} %{buildroot}/%{_datadir}/%{name}
install -m 644 %{SOURCE3} %{buildroot}/%{_datadir}/%{name}

%pre
getent group prometheus >/dev/null || groupadd -r prometheus
getent passwd prometheus >/dev/null || \
    useradd -r -g prometheus -d /var/lib/slurm_exporter -s /sbin/nologin \
    -c "Prometheus exporter user" prometheus
exit 0

%post
systemctl enable --now %{name}.service
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%files
%license %{_datadir}/%{name}/LICENSE
%doc %{_datadir}/%{name}/README.md
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%attr(755, prometheus, prometheus)/%{_sharedstatedir}/prometheus

%changelog
